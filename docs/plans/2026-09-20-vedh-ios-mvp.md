# vedh iOS MVP plan

**Status:** Implemented; core verified. Simulator and TestFlight verification are blocked on full Xcode, an iOS runtime, and signing access.

## Objective

Deliver a native SwiftUI client that covers the shortest complete vedh game loop:

1. authenticate;
2. create or join a Commander game;
3. see the table's current state;
4. change life and coordinate turn/priority state; and
5. share the game with another player.

The iOS client uses the existing Go/GraphQL API. It does not introduce a second backend or a mobile-only data model.

## MVP scope

| Area | Included now | Deferred |
|---|---|---|
| Account | Sign up, sign in/out, Keychain token storage | Password reset, account deletion |
| Lobby | Player's games, pull-to-refresh | Search, pagination UI, archived games |
| Create/join | Game ID or invite, commander search, decklist input, EDH defaults | Partner validation, format picker, deck validation |
| Live game | Polling, turn/phase/priority, life, stack and zone summaries | GraphQL WebSocket subscriptions |
| Actions | Own-life changes, pass priority, advance phase, claim win | Draw/mill/scry, tap/untap, card movement, stack resolution |
| Results | Finished state and win condition | History, charts, commander damage |
| Distribution | Generated Xcode project | TestFlight/App Store upload |

## Architecture

- `VedhCore`: Foundation-only data models, GraphQL transport, request builders, invite parsing, and game rules.
- `Vedh`: SwiftUI app and Keychain persistence.
- `VedhTests`: XCTest coverage against `VedhCore`.
- `VedhCoreCheck`: dependency-free checks that can run with Swift Command Line Tools when XCTest and Simulator SDKs are unavailable.
- XcodeGen owns project generation from `ios/project.yml`; generated project changes should be reproducible.

The app defaults to `https://api.vedh.xyz/graphql`. Use a local or staging endpoint for mutation-heavy automated tests so production game data is not polluted.

## Acceptance criteria

### Automated core checks

- [x] Invite parser accepts raw UUID-style IDs and vedh join/game URLs.
- [x] Invalid invite text is rejected.
- [x] Legacy phase names normalize to the server's current phase model.
- [x] Phase and priority rotation wrap correctly.
- [x] Current GraphQL response field names decode into Swift models.
- [x] Life mutation input preserves every card zone rather than overwriting the board with partial state.
- [x] GraphQL transport sends a Bearer token and decodes a response.
- [x] Xcode project and generated Info.plist pass property-list validation.
- [x] SwiftUI sources pass Swift parser validation.

### Simulator test matrix

Run against an isolated local or staging API with two disposable accounts.

| ID | Flow | Expected result |
|---|---|---|
| S1 | Launch without a stored session | Sign-in/create-account screen is shown. |
| S2 | Create an account | Session is saved and Games opens. Relaunch remains signed in. |
| S3 | Sign in with an invalid password | Server error is visible; no session is saved. |
| S4 | Search and select a commander | Matching cards appear and the selection persists in the form. |
| S5 | Create a game | Game appears in Games with turn 1, main phase, 40 life, and selected commander. |
| S6 | Share invite; join as account B | Account B joins and both players appear after the next refresh. |
| S7 | Change life by -1/+1/-5/+5 | Only the signed-in player's life changes; zone contents are preserved. |
| S8 | Pass priority A → B → A | Server and both clients show the same priority owner. Unauthorized controls are disabled. |
| S9 | Advance phase | Only the active player can advance; phase ordering and wrap are correct. |
| S10 | Claim win | Pending claim or finished state is rendered without losing board state. |
| S11 | Interrupt network, then restore | Stale status is explicit; polling recovers without relaunch. |
| S12 | VoiceOver and large text pass | All primary controls have useful labels and remain reachable. |
| S13 | Portrait/landscape on small and large iPhones | Life and turn state remain readable; no control is clipped. |

### Simulator commands

```sh
cd ios
xcodegen generate
xcodebuild \
  -project Vedh.xcodeproj \
  -scheme Vedh \
  -destination 'platform=iOS Simulator,name=iPhone 16 Pro' \
  test
```

After unit tests pass, execute S1–S13 manually while collecting screenshots for S5–S11 and preserving the Xcode test result bundle.

## TestFlight gate

Do not upload until all Simulator cases pass. TestFlight additionally requires:

1. an Apple Developer team and explicit signing selection;
2. an App Store Connect record for the final bundle ID;
3. final app icon, privacy copy, and support/privacy URLs;
4. an archive built with the release endpoint;
5. explicit approval before uploading externally; and
6. an internal test round on at least one physical iPhone before inviting external testers.

## Next implementation slice

After MVP verification, replace polling with authenticated GraphQL WebSockets and add reversible card-zone actions (draw, move, tap, resolve) using the same full-board-state preservation rule already enforced for life changes.
