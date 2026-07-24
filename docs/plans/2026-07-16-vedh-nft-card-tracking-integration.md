# vEDH NFT Card Tracking Integration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add an optional NFT-backed custom card system to vEDH so users can verify ownership of custom trading cards, import them into their vEDH card library, and track them in games without turning vEDH into a marketplace or rules engine.

**Architecture:** Keep gameplay state off-chain and keep vEDH as the canonical boardstate tracker. Add a wallet-link + ownership-verification layer, a custom-card metadata registry keyed by `chain_id + contract_address + token_id`, and a UI flow that lets users sync owned NFT cards into their personal library and use those cards in games. Ship EVM read-only ownership verification first; defer minting, listings, bidding, and cross-chain abstractions until the ownership loop is solid.

**Tech Stack:** Go, gqlgen GraphQL, PostgreSQL migrations in `persistence/migrations`, Vue 3, Pinia, Apollo GraphQL, Vitest, Go integration tests, optional EVM JSON-RPC via `go-ethereum` on the backend and `viem` on the frontend.

---

## Current grounded context

The current codebase is a strong fit for an ownership-tracking integration, but not yet for NFT-native data:

- vEDH is already a **format-agnostic boardstate tracker**, not a rules engine, per `docs/contextlog.md`
- auth is currently **username/password only** in:
  - `server/users.go`
  - `app/src/stores/auth.ts`
- GraphQL schema currently exposes:
  - `signup`, `login`
  - game creation/join/update
  - card search/query endpoints
  - no wallet, collection, or ownership types yet
- game state currently stores cards inline via the `Card` GraphQL type in `server/schema.graphql`
- board rendering is still heavily card-object driven in:
  - `app/src/views/BoardView.vue`
- frontend already has a working auth store and game store in:
  - `app/src/stores/auth.ts`
  - `app/src/stores/games.ts`
- backend already has a pattern for adding feature-specific GraphQL resolvers:
  - `server/formats.go`
  - `server/formats_api.go`
  - `server/formats_api_test.go`
- DB migrations run from `persistence/migrations` through `persistence/sql.go`

That means the safest first slice is:

1. add wallet linking to an existing vEDH account
2. verify NFT ownership off-chain via chain RPC reads
3. materialize owned custom cards into a vEDH-side card library
4. allow those cards to appear in vEDH board/deck flows

Do **not** start with on-chain minting, marketplace flows, or game-state writes to chain.

## Recommended product boundary

### Ship in v1

- Link one or more wallets to a vEDH account
- Verify ownership for one supported chain family (**EVM only** first)
- Support approved NFT collections that represent custom trading cards
- Import owned tokens as usable custom cards inside vEDH
- Preserve NFT provenance in card metadata and game logs
- Allow cards to remain usable in an active game even if ownership changes mid-game, while marking ownership drift after resync

### Explicitly do not ship in v1

- Minting from vEDH
- Marketplace listing/buy/sell/auction
- Trustless on-chain match state
- Cross-chain abstraction beyond EVM-compatible contracts
- Full arbitrary metadata editing in the browser
- Auto-enforcing card legality from the blockchain

## Hard decisions to make up front

These are real decisions, not filler. Block implementation at the first step if they remain fuzzy.

1. **Chain family**
   - Recommendation: EVM only first (`chain_id`, `contract_address`, `token_id`)
2. **Token standard support**
   - Recommendation: support both ERC-721 and ERC-1155 reads, but start test coverage with ERC-721 first
3. **Collection policy**
   - Recommendation: only whitelisted collections in v1; no arbitrary contract import
4. **Gameplay ownership rule**
   - Recommendation: ownership checked on import and optional resync, not every board interaction
5. **Custom-card metadata source of truth**
   - Recommendation: vEDH stores a normalized cached copy of NFT metadata plus a vEDH-specific play payload
6. **Auth model**
   - Recommendation: keep username/password auth, add wallet linking as a secondary identity method

---

### Task 1: Add the NFT domain model and migration scaffolding

**Files:**
- Create: `persistence/migrations/000019_nft_card_tracking.up.sql`
- Create: `persistence/migrations/000019_nft_card_tracking.down.sql`
- Create: `server/nft_models.go`
- Modify: `server/schema.graphql`
- Test: `server/graphql_httpserver_test.go`

**Step 1: Write the failing migration-oriented backend test**

Add a backend/API test that expects the new tables to exist after test DB setup.

Example assertion shape:

```go
func TestNFTTablesExist(t *testing.T) {
    db := testDB(t)
    for _, table := range []string{
        "wallet_identities",
        "nft_collections",
        "nft_card_templates",
        "nft_ownership_snapshots",
    } {
        var exists bool
        err := db.QueryRow(`
            SELECT EXISTS (
              SELECT 1 FROM information_schema.tables
              WHERE table_name = $1
            )
        `, table).Scan(&exists)
        require.NoError(t, err)
        require.True(t, exists, table)
    }
}
```

**Step 2: Run test to verify it fails**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run TestNFTTablesExist -v
```

Expected: FAIL because the schema does not exist yet.

**Step 3: Write the minimal schema and model layer**

Create migration tables for:

- `wallet_identities`
  - `id uuid primary key`
  - `user_id uuid not null`
  - `chain_id bigint not null`
  - `address text not null`
  - `verified_at timestamptz not null`
  - `created_at timestamptz not null default now()`
  - unique `(chain_id, address)`
- `wallet_link_challenges`
  - `id uuid primary key`
  - `user_id uuid not null`
  - `chain_id bigint not null`
  - `address text not null`
  - `nonce text not null`
  - `expires_at timestamptz not null`
  - `used_at timestamptz`
- `nft_collections`
  - `id uuid primary key`
  - `chain_id bigint not null`
  - `contract_address text not null`
  - `name text not null`
  - `symbol text`
  - `token_standard text not null`
  - `is_enabled boolean not null default true`
  - unique `(chain_id, contract_address)`
- `nft_card_templates`
  - `id uuid primary key`
  - `collection_id uuid not null`
  - `token_id text not null`
  - `name text not null`
  - `image_url text`
  - `metadata_url text`
  - `external_url text`
  - `vedh_card_payload jsonb not null`
  - `metadata_hash text`
  - unique `(collection_id, token_id)`
- `nft_ownership_snapshots`
  - `id uuid primary key`
  - `wallet_identity_id uuid not null`
  - `template_id uuid not null`
  - `quantity numeric not null default 1`
  - `owner_address text not null`
  - `synced_at timestamptz not null`
  - unique `(wallet_identity_id, template_id)`

Create `server/nft_models.go` with Go structs that mirror these records.

**Step 4: Run the targeted test**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run TestNFTTablesExist -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add persistence/migrations/000019_nft_card_tracking.* server/nft_models.go server/graphql_httpserver_test.go
git commit -m "feat: add NFT tracking schema scaffolding"
```

---

### Task 2: Add wallet-link challenge and verification flow

**Files:**
- Create: `server/wallet_auth.go`
- Create: `server/wallet_auth_test.go`
- Modify: `server/schema.graphql`
- Modify: `server/generated.go` (via gqlgen)
- Modify: `server/users.go`
- Modify: `app/src/stores/auth.ts`
- Create: `app/src/services/wallet.ts`
- Test: `server/wallet_auth_test.go`

**Step 1: Write the failing backend auth tests**

Add tests for:

- creating a wallet-link challenge for the authenticated user
- rejecting expired challenges
- rejecting reused nonces
- linking the wallet after a valid signature

Example assertion shape:

```go
func TestCreateWalletLinkChallenge(t *testing.T) {
    s := testAPI(t)
    challenge, err := s.CreateWalletLinkChallenge(authCtx("shakezula"), 1, "0xabc...")
    require.NoError(t, err)
    require.NotEmpty(t, challenge.Nonce)
    require.True(t, challenge.ExpiresAt.After(time.Now()))
}
```

**Step 2: Run tests to verify they fail**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run 'TestCreateWalletLinkChallenge|TestLinkWallet' -v
```

Expected: FAIL because no wallet auth resolver exists.

**Step 3: Add minimal schema and resolver support**

Extend `server/schema.graphql` with types/mutations like:

```graphql
type WalletIdentity {
  ID: String!
  ChainID: String!
  Address: String!
  VerifiedAt: Time!
}

type WalletLinkChallenge {
  Address: String!
  ChainID: String!
  Nonce: String!
  Message: String!
  ExpiresAt: Time!
}

extend type Mutation {
  createWalletLinkChallenge(chainID: String!, address: String!): WalletLinkChallenge!
  linkWallet(chainID: String!, address: String!, signature: String!): WalletIdentity!
}
```

Implement backend behavior in `server/wallet_auth.go`:

- require existing vEDH auth
- normalize address to lowercase checksum-insensitive storage key
- create short-lived nonce challenge
- verify EVM personal-sign/SIWE-style signature server-side
- persist `wallet_identities`
- invalidate used challenge

Implement a tiny frontend helper in `app/src/services/wallet.ts` using `window.ethereum` first; do not add WalletConnect yet.

**Step 4: Regenerate gqlgen**

Run:

```bash
cd /root/.openclaw/workspace/vedh
make generate
```

Expected: generated files update cleanly.

**Step 5: Run backend tests**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run 'TestCreateWalletLinkChallenge|TestLinkWallet' -v
```

Expected: PASS.

**Step 6: Commit**

```bash
git add server/schema.graphql server/wallet_auth.go server/wallet_auth_test.go server/generated.go server/users.go app/src/stores/auth.ts app/src/services/wallet.ts
git commit -m "feat: add wallet linking flow"
```

---

### Task 3: Add approved NFT collection registry and EVM read provider

**Files:**
- Create: `server/nft_provider.go`
- Create: `server/nft_provider_evm.go`
- Create: `server/nft_provider_evm_test.go`
- Create: `server/nft_collections.go`
- Modify: `server/schema.graphql`
- Test: `server/nft_provider_evm_test.go`

**Step 1: Write the failing provider tests**

Add tests for:

- reading ERC-721 owner for a token
- reading ERC-1155 balance for a token
- rejecting an unsupported token standard
- normalizing contract addresses

Example assertion shape:

```go
func TestEVMProviderOwnerOf721(t *testing.T) {
    provider := newMockEVMProvider(t)
    owner, qty, err := provider.LookupOwnership(context.Background(), CollectionRef{
        ChainID: 1,
        ContractAddress: "0x1234...",
        TokenStandard: "ERC721",
    }, "42", "0xabcd...")
    require.NoError(t, err)
    require.Equal(t, "0xabcd...", owner)
    require.Equal(t, 1, qty)
}
```

**Step 2: Run tests to verify they fail**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run 'TestEVMProvider' -v
```

Expected: FAIL because provider layer does not exist.

**Step 3: Implement minimal provider abstraction**

Create interface in `server/nft_provider.go`:

```go
type NFTProvider interface {
    LookupOwnership(ctx context.Context, collection CollectionRef, tokenID string, address string) (owner string, quantity int, err error)
    FetchMetadata(ctx context.Context, collection CollectionRef, tokenID string) (*NFTMetadata, error)
}
```

Implement `server/nft_provider_evm.go` with:

- EVM RPC client
- ERC-721 `ownerOf`
- ERC-1155 `balanceOf`
- metadata URI fetch for whitelisted collections

Create `server/nft_collections.go` for CRUD/lookup of approved collections.

**Step 4: Run provider tests**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run 'TestEVMProvider' -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add server/nft_provider.go server/nft_provider_evm.go server/nft_provider_evm_test.go server/nft_collections.go
git commit -m "feat: add EVM NFT ownership provider"
```

---

### Task 4: Materialize NFT metadata into vEDH custom-card templates

**Files:**
- Create: `server/nft_sync.go`
- Create: `server/nft_sync_test.go`
- Modify: `server/schema.graphql`
- Modify: `server/formats.go` (only if custom-card payload needs format helpers)
- Test: `server/nft_sync_test.go`

**Step 1: Write the failing sync tests**

Add tests for:

- syncing a wallet-owned token creates a `nft_card_templates` record
- syncing again updates metadata instead of duplicating
- unowned token is removed or quantity zeroed in `nft_ownership_snapshots`
- malformed metadata is rejected cleanly

**Step 2: Run tests to verify they fail**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run 'TestSyncNFTCards' -v
```

Expected: FAIL because no sync flow exists.

**Step 3: Implement minimal sync behavior**

Create mutation in schema:

```graphql
extend type Mutation {
  syncMyNFTCards: [NFTCard!]!
}
```

Add types:

```graphql
type NFTCard {
  ID: String!
  ChainID: String!
  ContractAddress: String!
  TokenID: String!
  Name: String!
  ImageURL: String
  MetadataURL: String
  Quantity: Int!
  OwnershipVerifiedAt: Time!
}
```

Implement `server/nft_sync.go` to:

- load linked wallets for current user
- scan approved collections for owned token IDs
- fetch metadata
- normalize metadata into a vEDH payload (`name`, `image_url`, optional `types`, optional custom stats text)
- upsert `nft_card_templates`
- upsert `nft_ownership_snapshots`

**Step 4: Run sync tests**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run 'TestSyncNFTCards' -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add server/schema.graphql server/nft_sync.go server/nft_sync_test.go
git commit -m "feat: sync owned NFT cards into vedh templates"
```

---

### Task 5: Expose NFT-backed cards through GraphQL queries and search

**Files:**
- Modify: `server/schema.graphql`
- Modify: `server/formats_api.go` or create `server/nft_api.go`
- Modify: `server/formats_api_test.go` or create `server/nft_api_test.go`
- Modify: `app/src/graphql/queries.ts`
- Modify: `app/src/stores/games.ts`
- Test: `server/nft_api_test.go`

**Step 1: Write the failing API tests**

Add tests for:

- `myNFTCards` returns linked-user owned cards only
- `searchCustomCards(query: ...)` returns NFT-backed custom cards
- unlinked users get empty results, not other users’ cards

**Step 2: Run tests to verify they fail**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run 'TestMyNFTCards|TestSearchCustomCards' -v
```

Expected: FAIL because queries do not exist.

**Step 3: Implement GraphQL read layer**

Add schema:

```graphql
extend type Query {
  myNFTCards: [NFTCard!]!
  searchCustomCards(query: String!): [NFTCard!]!
}
```

Implement resolvers in `server/nft_api.go`:

- filter by authenticated user ownership snapshots
- join template + collection data
- optionally search name and token ID text

Add matching client queries in `app/src/graphql/queries.ts`.

**Step 4: Run backend tests**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run 'TestMyNFTCards|TestSearchCustomCards' -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add server/schema.graphql server/nft_api.go server/nft_api_test.go app/src/graphql/queries.ts app/src/stores/games.ts
git commit -m "feat: expose NFT card queries"
```

---

### Task 6: Let the board and deck flows carry NFT provenance safely

**Files:**
- Modify: `server/schema.graphql`
- Modify: `server/games.go`
- Modify: `server/searchall_integration_test.go`
- Modify: `app/src/components/games/FormCreateGame.vue`
- Modify: `app/src/views/BoardView.vue`
- Modify: `app/src/stores/games.ts`
- Test: `app/__tests__/FormCreateGame.integration.spec.ts`
- Test: `app/e2e/create-and-join-game.spec.ts`

**Step 1: Write the failing frontend and backend tests**

Add tests asserting:

- a selected NFT-backed card can be included in deck import payload
- board view renders NFT card art via `ImageURL`/metadata fallback when Scryfall art is absent
- joining and loading a game preserves `chain_id`, `contract_address`, and `token_id`

**Step 2: Run tests to verify they fail**

Run:

```bash
cd /root/.openclaw/workspace/vedh/app
npm test -- FormCreateGame.integration.spec.ts
```

and

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run 'TestGamePersistsNFTCardMetadata' -v
```

Expected: FAIL because card provenance fields are missing.

**Step 3: Add minimal provenance fields to `Card`**

Extend `Card` / `InputCard` in `server/schema.graphql` with:

```graphql
ChainID: String
ContractAddress: String
TokenID: String
MetadataURL: String
ImageURL: String
CardSource: String
```

Persist those values through game create/update/join flows in `server/games.go`.

Update `BoardView.vue` image resolution logic to:

1. prefer `ImageURL`
2. fall back to current Scryfall/lookup image
3. show placeholder if both absent

**Step 4: Run tests**

Run:

```bash
cd /root/.openclaw/workspace/vedh
make test-api
cd /root/.openclaw/workspace/vedh/app
npm test
```

Expected: PASS.

**Step 5: Commit**

```bash
git add server/schema.graphql server/games.go app/src/components/games/FormCreateGame.vue app/src/views/BoardView.vue app/src/stores/games.ts app/__tests__/FormCreateGame.integration.spec.ts app/e2e/create-and-join-game.spec.ts
git commit -m "feat: preserve NFT card provenance in gameplay"
```

---

### Task 7: Add a user-facing wallet + NFT library screen

**Files:**
- Create: `app/src/views/ProfileView.vue`
- Modify: `app/src/router/index.ts`
- Modify: `app/src/stores/auth.ts`
- Create: `app/src/stores/nftCards.ts`
- Create: `app/__tests__/NFTLibraryView.spec.ts`
- Modify: `app/src/graphql/queries.ts`
- Modify: `app/src/graphql/mutations.ts`

**Step 1: Write the failing frontend test**

Add a Vue test that expects the profile/library screen to:

- show linked wallets
- offer a “Link wallet” button
- offer a “Sync NFT cards” button
- list imported cards

**Step 2: Run test to verify it fails**

Run:

```bash
cd /root/.openclaw/workspace/vedh/app
npm test -- NFTLibraryView.spec.ts
```

Expected: FAIL because the route/store/view do not exist.

**Step 3: Build the minimal UI**

Create `ProfileView.vue` with sections:

- Account
- Linked wallets
- Owned NFT cards

Add a dedicated Pinia store `app/src/stores/nftCards.ts` for:

- `linkedWallets`
- `cards`
- `syncing`
- `errorMessage`

Add route:

- `/profile`

**Step 4: Run the frontend test**

Run:

```bash
cd /root/.openclaw/workspace/vedh/app
npm test -- NFTLibraryView.spec.ts
npm run type-check
```

Expected: PASS.

**Step 5: Commit**

```bash
git add app/src/views/ProfileView.vue app/src/router/index.ts app/src/stores/nftCards.ts app/src/stores/auth.ts app/src/graphql/queries.ts app/src/graphql/mutations.ts app/__tests__/NFTLibraryView.spec.ts
git commit -m "feat: add wallet and NFT library UI"
```

---

### Task 8: Add ownership-drift handling and game-log provenance

**Files:**
- Modify: `server/games.go`
- Modify: `server/schema.graphql`
- Modify: `server/graphql.go` or relevant logging helpers
- Test: `server/games_test.go`
- Test: `server/graphql_httpserver_test.go`

**Step 1: Write the failing tests**

Add tests for:

- logging provenance when an NFT-backed card enters a game
- marking ownership drift after a resync if the card is no longer owned
- not removing the card from an already-running game automatically

**Step 2: Run tests to verify they fail**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run 'TestNFTProvenanceLogging|TestOwnershipDrift' -v
```

Expected: FAIL because game-log integration does not exist.

**Step 3: Implement minimal logging + drift status**

Add fields to NFT GraphQL read type if needed:

```graphql
OwnershipStatus: String!
LastVerifiedAt: Time!
```

Add game-log payload entries when:

- NFT card added to a board/deck
- ownership drift detected on sync

Recommended statuses:

- `VERIFIED`
- `STALE`
- `NOT_OWNED_ANYMORE`

**Step 4: Run backend tests**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run 'TestNFTProvenanceLogging|TestOwnershipDrift' -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add server/games.go server/schema.graphql server/graphql_httpserver_test.go server/games_test.go
git commit -m "feat: log NFT provenance and ownership drift"
```

---

### Task 9: Add smoke coverage and operator docs

**Files:**
- Modify: `README.md`
- Create: `docs/plans/2026-07-16-vedh-nft-rollout-checklist.md`
- Modify: `app/e2e/create-and-join-game.spec.ts`
- Modify: `tools/smoke/src/main.rs` (only if extending the Rust smoke path is worth it)
- Test: existing Go/Vue/Playwright suites

**Step 1: Write the failing doc/test TODO checks**

Create one failing smoke or E2E assertion that:

- links a wallet in mocked mode
- syncs one NFT-backed card
- creates a game with that card visible in a board zone

**Step 2: Run test to verify it fails**

Run:

```bash
cd /root/.openclaw/workspace/vedh/app
npm run test:e2e -- --grep "NFT"
```

Expected: FAIL because end-to-end NFT flow is not wired.

**Step 3: Add rollout docs and smoke notes**

Update `README.md` with:

- required env vars for EVM RPC
- how wallet-link auth works
- how collection whitelisting works
- how to resync NFT cards

Create rollout checklist doc with:

- RPC provider setup
- collection whitelist seed
- smoke steps
- ownership-drift verification step
- rollback plan

**Step 4: Run final proof commands**

Run:

```bash
cd /root/.openclaw/workspace/vedh
make test-api
cd /root/.openclaw/workspace/vedh/app
npm test
npm run test:e2e
```

Expected: PASS.

**Step 5: Commit**

```bash
git add README.md docs/plans/2026-07-16-vedh-nft-rollout-checklist.md app/e2e/create-and-join-game.spec.ts tools/smoke/src/main.rs
git commit -m "docs: add NFT rollout and smoke coverage"
```

---

## Suggested env vars for v1

Backend:

```bash
EVM_RPC_URL="https://..."
NFT_SUPPORTED_CHAIN_IDS="1,8453"
NFT_COLLECTION_ALLOWLIST="1:0xabc...,8453:0xdef..."
NFT_CHALLENGE_TTL_SECONDS="300"
```

Frontend:

```bash
VITE_ENABLE_NFT_CARDS="true"
```

## Data-shape recommendation for custom cards

Normalize NFT metadata into a vEDH payload like:

```json
{
  "name": "Custom Shock Drake",
  "types": "Creature — Drake",
  "text": "Flying\nWhen this enters, deal 2 damage to any target.",
  "power": "2",
  "toughness": "1",
  "image_url": "https://...",
  "colors": "U,R",
  "source": "NFT"
}
```

That keeps game rendering simple because the board already expects card-ish fields.

## Biggest risks

1. **Wallet auth complexity**
   - Mitigation: keep existing auth, only add linking
2. **Untrusted NFT metadata quality**
   - Mitigation: whitelist collections and normalize metadata server-side
3. **Board UI assumptions around Scryfall/MTG cards**
   - Mitigation: add explicit `ImageURL` + `CardSource` before trying to reuse search flows everywhere
4. **Scope explosion into marketplace work**
   - Mitigation: explicitly reject mint/list/sell in v1
5. **Ownership drift during a match**
   - Mitigation: treat ownership as import-time verification plus resync status, not hard real-time gameplay enforcement

## Recommended first implementation slice

If you want the fastest path to something real, do only these tasks first:

1. Task 1 — schema/migrations
2. Task 2 — wallet linking
3. Task 3 — EVM provider
4. Task 4 — sync owned NFT cards
5. Task 7 — basic wallet/library UI

That gives you a usable ownership-tracked custom-card library before touching the board UX too deeply.

## Recommendation

My recommendation is **EVM read-only, whitelisted collections, linked-wallet auth, no minting** for v1.

That gets vEDH into NFTs in a way that actually matches the product: tracking, identity, provenance, and play usage — not speculative marketplace baggage.
