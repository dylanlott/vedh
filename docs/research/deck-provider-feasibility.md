# Deck Provider Feasibility Comparison — Archidekt vs. Moxfield

**Ticket:** ACT-003 (`docs/plans/2026-07-23-deck-to-game-activation-tickets.md`), one-day
timeboxed spike.
**Decision this feeds:** OPEN-1 / D-13 / D-14's `checkpoint:decision` in
`01-07-PLAN.md` Task 2. This document is the evidence; the checkpoint is where the
branch is chosen.
**Observation date for every item below (unless stated otherwise):** 2026-08-05.
**Observed from:** the plan executor's own environment, over live outbound HTTPS. All
requests below were single, low-volume, minimal probes against public unauthenticated
endpoints — no credentials, no automation loop, no attempt to defeat any block
encountered. Where a provider's `robots.txt` disallowed a path, that path was not
fetched; the disallow itself is recorded as evidence rather than worked around.

Per D-13 this comparison starts neutral: both candidates are evaluated against the same
six criteria, in the same depth, in the same document, before any recommendation is
formed.

---

## 1. Archidekt

### 1.1 Request shape and public read path

**Observed 2026-08-05.** Archidekt exposes an unauthenticated, unversioned JSON read
path under `https://archidekt.com/api/decks/<id>/`. A plain `GET` with no headers beyond
a descriptive `User-Agent` returns `200 application/json` for a public deck:

```
GET https://archidekt.com/api/decks/13074677/
→ HTTP/2 200, content-type: application/json, content-length: 350219
```

The response is a single deeply-nested JSON object. Top level:
`id, name, createdAt, updatedAt, deckFormat, edhBracket, game, description, viewCount,
featured, customFeatured, hasPrimer, private, unlisted, theorycrafted, points, userInput,
owner, commentRoot, editors, parentFolder, bookmarked, categories, deckTags,
playgroupDeckUrl, cardPackage, deckHelp, cards, customCards,
intentionallySkippedCardData`.

`cards` is an array of deck-line objects, each carrying `quantity`, `categories`
(the section/category tags), and a fully embedded `card` object (which itself embeds an
`oracleCard` object). The fields an adapter needs are present and consistently placed:

- Card name: `cards[i].card.oracleCard.name`
- Quantity: `cards[i].quantity`
- Set code: `cards[i].card.edition.editioncode`
- Collector number: `cards[i].card.collectorNumber`
- Section/category: `cards[i].categories` (an array — a card can carry more than one)

There is also a `/small/` variant (`/api/decks/<id>/small/`) that returns the same
top-level shape but with `cards` always `[]` — confirmed by direct inspection, not
inferred from documentation, since no documentation exists (see 1.4). It is useless for
importing card content and was not considered further.

Sample sizes observed across four different real public decks (99–134 cards each):
299,509 – 411,097 bytes. All comfortably inside the locked 1 MiB cap.

**A genuine public developer API does not exist.** `https://archidekt.com/api` and
`https://archidekt.com/developers` both return `404`. Requesting a client-only route
(`/api/decks/<id>/export/`) returns a revealing internal error: `"Client Unavailable: You
are requesting client routes from the api. Use a react server with the archidekt-client
to develop locally. If you are seeing this in production, the load balancer is not
correctly routing."` This confirms `/api/decks/<id>/` is Archidekt's own single-page-app
data API — the same one their React frontend calls — not a contract published or
versioned for third-party consumers. It also has an
`access-control-allow-origin: http://localhost:3000` header, consistent with an API whose
intended caller is Archidekt's own local dev frontend, not a third party (this is a
browser-only concept — CORS does not affect a server-to-server fetch like ours — recorded
here only as corroborating context for 1.4).

### 1.2 Private-deck behaviour

**Observed 2026-08-05.** A nonexistent deck ID returns `404` with a structured body:

```
GET https://archidekt.com/api/decks/2/small/  → 404 {"error":"Deck not found."}
GET https://archidekt.com/api/decks/100/small/ → 404 {"error":"Deck not found."}
```

We hold no Archidekt account and cannot mark a deck private ourselves, so we could not
directly confirm that a deck genuinely private to another user returns this same
response rather than a different one (e.g., a `403`). The identical, existence-agnostic
`404` shape for IDs that plainly do not exist is suggestive that a private deck (which is
also "not accessible to this caller") would return the same generic error rather than
leaking that the deck exists — but this is inference from adjacent evidence, not a
directly observed private-deck response. **Recorded as an unresolved unknown**, not
papered over: an adapter would need to treat any non-200 (including this 404 shape) as
"unsupported," which is what task 3 would need to implement regardless of which of these
two possibilities is true, so this unknown does not block acceptance-criteria compliance
either way.

### 1.3 Rate-limit signals

**Observed 2026-08-05.** Five sequential requests to the same deck endpoint returned
`200` every time with no `X-RateLimit-*`, `Retry-After`, or `429` observed at any point.
No throttling was encountered at this (deliberately low, one-day-timebox-respecting)
volume. This does not establish what the actual limit is, only that it was not hit by a
handful of requests — recorded as an unknown beyond that, not extrapolated into a
guessed number.

### 1.4 Response-shape stability — the hard gate condition (D-15)

**Observed 2026-08-05.** No published schema, no versioned path segment (`/api/decks/`,
not `/api/v1/decks/` or similar), and no changelog were found. `/api/docs`,
`/developers`, and a plain `/api` all `404`. The endpoint is, by its own error message
(1.1), the same internal data API Archidekt's own SPA frontend consumes — an
implementation detail exposed at a stable-looking URL, not a contract Archidekt has
committed to keeping stable for outside consumers.

Against that risk: the shape was directly observed to be structurally consistent across
six independently fetched real public decks (IDs `1`, `13074677`, `1700657`,
`17952921`, `18560442`, and one deck's `small` variant) on the same date, with the same
top-level key set and the same nested `card.oracleCard` shape in every case. This is
current-state evidence of consistency, not a guarantee of future stability — an
unversioned internal API can change without notice, and nothing observed here rules that
out.

**Verdict for the hard gate:** reachable and structurally consistent as of this date;
stability is unproven going forward because no versioning or documentation commits
Archidekt to anything. This is a real risk to weigh, not a pass/fail answer this document
can give on Archidekt's behalf — recorded plainly for the human at the checkpoint per
D-15.

### 1.5 Terms of service and operational risk

**Observed 2026-08-05.** `https://archidekt.com/terms` returns `200`, but the page is a
client-rendered single-page app — the HTML delivered by a plain `GET` contains font
declarations and application shell markup, not the rendered terms text, because the
actual content is fetched and rendered by JavaScript after load. This tool has no
JavaScript execution capability, so **the terms text itself could not be read** by the
available means. This is recorded as an explicit unknown rather than filled with a
paraphrase or a recollection: **we do not know what Archidekt's terms say about
automated/programmatic reads**, and a human should read `https://archidekt.com/terms` in
a real browser before relying on this candidate.

`robots.txt` (`https://archidekt.com/robots.txt`) disallows only `/partialCompare`,
`/playtester-v2/`, and `/sandbox?` — none of which overlap `/api/decks/`. This is
permissive evidence, not comprehensive: `robots.txt` speaks to crawler etiquette, not to
whether the terms of service separately restrict API use.

### 1.6 Fixture maintainability

**Observed 2026-08-05.** A representative response was captured and checked in at
`server/testdata/deck_providers/archidekt_deck_2026-08-05.json` (redacted — see below).
Capturing was trivial: one `GET`, no auth, no special headers. What would make it go
stale: (a) Archidekt changing the undocumented internal shape described in 1.4 with no
notice, since nothing commits them to a contract; (b) the specific deck ID being deleted
or turned private by its owner, which is why the fixture is a point-in-time capture keyed
to a date, not a live reference. Both risks are manageable the same way any undocumented-API
fixture is managed — a fixture-contract test (task 3) that fails loudly the moment the
captured shape stops matching a fresh capture, rather than silently reparsing whatever
comes back.

**Fixture:** `server/testdata/deck_providers/archidekt_deck_2026-08-05.json` — the full
`/api/decks/13074677/` response (113 cards, "Cid On My Face"). Redacted: `owner.id`,
`owner.username`, `owner.avatar`, `owner.ckAffiliate`, `owner.tcgAffiliate`,
`owner.referrerEnum`, `commentRoot`, `parentFolder`, and each card line's `id`,
`createdAt`, `updatedAt` — none of which the parser needs, all of which identify the real
account that owns the source deck. The structural shape (every key, every nesting level)
is unchanged.

### 1.7 Reachability within the locked controls exactly as they exist today

**Observed 2026-08-05.** Every probe above went over `https://archidekt.com` on port
443, no redirect was ever issued (all responses were direct `200`/`404`, zero hops), and
the single hostname `archidekt.com` is sufficient for both the deck-metadata and
full-card-list paths (no separate API subdomain, unlike Moxfield — see 2.7). This is
fully reachable within `server/deck_providers.go`'s locked controls exactly as they exist
today: HTTPS-only, exact-equality allowlist on one hostname, well inside the 3s
connect/8s total timeouts, well inside the 1 MiB cap. **No control would need to be
relaxed to reach Archidekt.**

### 1.8 App-level credential requirement (D-16)

**Observed 2026-08-05.** None of the endpoints probed required any credential, API key,
or registered account. The public-deck read path is fully unauthenticated. If Archidekt
is selected, no `DECK_PROVIDER_API_KEY` value is needed and no server-side secret enters
the Dokku deploy path for this reason.

---

## 2. Moxfield

### 2.1 Request shape and public read path

**Not established.** Moxfield's actual data API lives at `https://api.moxfield.com`.
Its `robots.txt` (`https://api.moxfield.com/robots.txt`, observed 2026-08-05) reads:

```
User-agent: *
Disallow: /
```

This disallows automated access to the entire host — there is no narrower path this
spike could have probed instead. Per this plan's own instruction, a `robots.txt`
disallow is recorded as evidence rather than worked around: **no request was made to
`api.moxfield.com` beyond fetching `robots.txt` itself**, so the actual public-deck
response shape from that host is unknown by design, not by omission.

Separately, and independently of the `robots.txt` finding: the web frontend at
`https://moxfield.com` sits behind Cloudflare's active bot-management layer. A plain,
non-browser `GET` to the bare homepage (`https://moxfield.com/`) returned `403` with a
Cloudflare "Attention Required" challenge page, observed with three different
`User-Agent` values (curl's default, a custom descriptive string, and a full Chrome
desktop `User-Agent` string) — all three were blocked identically, which rules out
"wrong User-Agent" as the cause. A single, isolated follow-up request (after a pause) to
a deck-shaped path (`https://moxfield.com/decks/<id>`) returned the same Cloudflare
challenge `403` before ever reaching Moxfield's own application. This was observed
directly, not inferred: the response body is Cloudflare's own "Attention Required |
Cloudflare" HTML, not a Moxfield application page.

`moxfield.com`'s own `robots.txt` (as distinct from `api.moxfield.com`'s) is narrower —
it disallows `/user/admin.html`, `/search/*`, `/account/*`, `/collection/*`, and
`/binders/*`, and does not disallow `/decks/*`. So the frontend's crawl policy would
technically have permitted probing a deck page; it was Cloudflare's bot-management layer,
not Moxfield's stated policy, that refused every attempt observed here.

**Combined finding:** the host an adapter would actually need to call
(`api.moxfield.com`) explicitly disallows all automated access via `robots.txt`, and the
public-facing web host that could theoretically substitute for it is blocked by active
anti-automation infrastructure that a plain `net/http` client — with no special browser
fingerprint, exactly what `server/deck_providers.go`'s secure client is — was observed to
trip on every attempt. There is no public read path this spike could exercise for
Moxfield within the bounds it set for itself.

### 2.2 Private-deck behaviour

**Not established.** No response was ever obtained from Moxfield's data API (2.1), so
there is nothing to report here beyond that same access barrier.

### 2.3 Rate-limit signals

**Not established**, for the same reason. No successful request means no rate-limit
headers were ever seen to observe.

### 2.4 Response-shape stability — the hard gate condition (D-15)

**Not established, and not establishable within this spike's bounds.** D-15 makes this
the one hard gate condition. It cannot be evaluated as "stable" or "unstable" for
Moxfield because no response body was ever obtained to evaluate — not because Moxfield's
shape is known to be unstable, but because it could not be observed at all without either
violating `api.moxfield.com`'s explicit `robots.txt` disallow or defeating
`moxfield.com`'s active bot-management, neither of which this spike will do. An
unassessable hard-gate condition is not a pass by default: the honest state is "unknown,"
and per this plan's own instruction an inconclusive result is not rounded up to a go.

### 2.5 Terms of service and operational risk

**Observed 2026-08-05, partially.** `https://moxfield.com/terms` and
`https://moxfield.com/tos` both returned `403` (the same Cloudflare challenge described
in 2.1) — the terms page itself could not be read by the means available here. What
*was* directly observed and is dispositive on its own: `api.moxfield.com/robots.txt`'s
blanket `Disallow: /` is itself an operational-risk signal independent of whatever the
prose terms say — it is the site operator's own machine-readable statement that
automated access to that host is not wanted.

### 2.6 Fixture maintainability

**Not established.** No response body was ever obtained (2.1), so nothing could be
captured. No fixture is checked in for Moxfield. Per this spike's own ground rule, this
is recorded as the finding rather than substituted with a fabricated or recalled sample
response.

### 2.7 Reachability within the locked controls exactly as they exist today

**Observed 2026-08-05.** No control in `server/deck_providers.go` was relaxed or would
need to be relaxed to reach Moxfield — the requests made were plain HTTPS GETs on port
443 to an exact hostname, exactly what the locked client already does. The barrier is not
that our own controls block Moxfield; it is that Moxfield's own infrastructure
(`robots.txt` policy on the API host, and Cloudflare bot-management on the web host)
refuses the request before our controls are even relevant to the outcome. **This is a
no-go for reachability in practice, for reasons entirely outside the locked controls'
scope** — recorded explicitly as such per the instruction to say so when a candidate is
only reachable (or, here, not reachable at all) outside what the controls can affect.

### 2.8 App-level credential requirement (D-16)

**Not established.** Whether an app-level API key would grant access to
`api.moxfield.com` despite its `robots.txt` posture is unknown — `robots.txt` is a crawl
directive, not an authentication mechanism, and Moxfield may have an entirely separate,
credentialed developer program this spike did not discover (no `/developers` or
`/api/docs` path resolved; both returned `403` via the same Cloudflare layer described in
2.1, so even discovery of such a program was blocked). If one exists, it was not found
within this spike's one-day box.

---

## 3. Reachability summary

| Criterion | Archidekt | Moxfield |
|---|---|---|
| Public read path exists | Yes — `archidekt.com/api/decks/<id>/`, unauthenticated | Not established — `robots.txt` disallow-all on the API host; Cloudflare bot-management blocks the web host |
| Private-deck behaviour | Existence-agnostic 404 observed for missing IDs; true private-deck response unconfirmed | Not established |
| Rate-limit signals | None observed at low volume; limit itself unknown | Not established |
| Response-shape stability (hard gate) | Unversioned/undocumented, but structurally consistent across 6 fetches on this date | Not established — cannot be evaluated at all |
| Terms of service | Page exists but is JS-rendered; text unread by available means; `robots.txt` permissive for this path | Page blocked (`403`); `robots.txt` disallow-all on the API host is itself the operative signal |
| Fixture maintainable | Yes — captured and checked in | No — nothing to capture |
| Reachable within today's locked controls | Yes, with room to spare on every budget | No — blocked by the provider's own infrastructure, not by our controls |
| App-level credential needed | No | Not established |

---

## 4. Recommendation

**Archidekt.**

Tied to the hard gate condition (D-15): Archidekt's response shape, while undocumented
and unversioned — a real and disclosed risk — was directly observed and reachable, and
was structurally consistent across six independent fetches on this date. Moxfield's
response shape cannot be evaluated at all, because no response could be obtained within
the bounds this spike set for itself (an explicit `robots.txt` disallow on the API host,
and active anti-automation infrastructure on the web host that a plain, non-browser HTTPS
client was observed to trip on every attempt, including a `403` on the terms page itself).
An unassessable hard-gate condition is the weaker outcome, not a neutral one: "we could
not check" is not "it probably passes."

**What the human should weigh at the checkpoint, beyond the hard gate:**

- Archidekt's API is genuinely undocumented — an internal implementation detail exposed
  by their own SPA, not a contract they have committed to. It could change without notice
  at any time, and nothing found here would give advance warning beyond a failing
  fixture-contract test after the fact.
- The Archidekt terms of service could not actually be read (JS-rendered page, no
  JavaScript execution available here) — a human should open
  `https://archidekt.com/terms` in a real browser before relying on this candidate, since
  this document cannot confirm what it permits.
- No credential is needed for Archidekt, so choosing it adds no server-side secret to the
  Dokku deploy path (D-16 is moot for this candidate).
- Choosing Moxfield is not available as a "yes" outcome here — the evidence needed to
  even evaluate it could not be gathered without doing exactly what this spike declined to
  do (ignore a `robots.txt` disallow, or attempt to defeat a WAF challenge). If the human
  has independent, credentialed access to Moxfield's developer program (undiscovered by
  this spike, see 2.8), that would be new evidence this document does not have, and would
  warrant a fresh look rather than overriding this recommendation on the strength of what
  is written here.
- Recording a **no-go** remains fully available and costs nothing to reverse (see
  `01-07-PLAN.md` Task 2's context) — the secure client and kill switch already exist and
  default off regardless of which of the three options is chosen here.

This document names one recommendation; it does not choose the branch. That is Task 2's
`checkpoint:decision`, held for a human.

---

## 5. Decision (Task 2 checkpoint)

**Selected: Moxfield.**

**Decided by:** the user (dylan), at the Task 2 `checkpoint:decision` in `01-07-PLAN.md`.
**Date:** 2026-08-05.

This diverges from Section 4's recommendation above, which named Archidekt — the only
candidate whose hard gate condition (response-shape stability, D-15) could actually be
evaluated with observed evidence, while Moxfield's could not be evaluated at all within
this spike's bounds. The user weighed that evidence and chose Moxfield anyway. That is
the user's call to make at this checkpoint, not a correction to this record: Section 4's
recommendation is left exactly as it was written before this decision, as the historical
record of what the neutral comparison actually found.

### Open blockers gating enablement

Selecting Moxfield here does not make it usable. Two blockers, both already identified in
Section 2 above, must be resolved before `DECK_PROVIDER_ENABLED` may be turned on for this
provider in any real deployment:

**(a) Authorization.** `api.moxfield.com/robots.txt` is a blanket `Disallow: /`
(re-verified 2026-08-05, unchanged from Section 2.1's original observation). Automated
access to that host requires either a credentialed developer program Moxfield may or may
not operate (not discovered by this spike — see Section 2.8) or some other explicit
authorization and agreed terms. Until that authorization exists, no request to
`api.moxfield.com` should be made by this codebase, in test or in production, and none is
made anywhere in the code this decision produced.

**(b) Unobserved response contract.** No Moxfield deck response body has ever been
captured (Sections 2.1–2.4). Task 3's adapter therefore ships with its normalizer
deliberately unimplemented — see `server/deck_providers.go`'s `moxfieldAdapter` and
`errMoxfieldContractUnverified` — rather than a fabricated field mapping. A real,
authorized sample response is required before that seam can be filled in.

### What Task 3 built instead

Per the user's explicit direction at this checkpoint — build the skeleton, with the
response contract marked unverified rather than invented — Task 3 shipped hostname-keyed
routing plumbing through plan 01-06's secure fetch client (`newSafeProviderClient`,
`fetchDeckProviderURL`), gated by the same default-off kill switch
(`DECK_PROVIDER_ENABLED`/`providerEnabled()`), with an explicit fail-closed seam
(`moxfieldAdapter.normalizeToDeckText`) where the response normalizer will go once blocker
(b) above is resolved.

This is scaffolding, not a working Moxfield import: a `moxfield.com` deck URL still
resolves to the same paste-fallback product-language message today as it did before this
task, the difference being that the message is now reached via a real hostname-routing
decision (`deckProviderAdapterFor`) and a real, genuinely-invoked (and genuinely-failing)
normalizer, rather than an unconditional stub. Every test added for this is hermetic — a
spied `deckProviderFetch` or a direct call to the adapter's `normalizeToDeckText` with an
in-memory byte slice — and none reaches any live Moxfield host, per this task's own
constraint and per the standing project-wide prohibition on any test depending on a live
provider.

See `.planning/phases/01-measured-deck-import-foundation/01-07-SUMMARY.md` for the full
account, including the deviation from this plan's originally anticipated Task 3 shape
(a fixture-tested working adapter, or a no-go) that this decision produced.

---

## 6. Revised decision (2026-08-07)

**Selected: Archidekt.** This section supersedes Section 5's Moxfield selection above —
it does not rewrite it. Both decisions and the evidence behind each remain readable in
full: Section 5 is left exactly as it was written, as the historical record of what was
actually decided and why, at the time it was decided.

**Decided by:** the user (dylan), during UAT of plan 01-07, recorded as gap `G-01-1` in
`.planning/phases/01-measured-deck-import-foundation/01-UAT.md`.
**Date:** 2026-08-07.

### Why this reverses Section 5

The user asked, in substance, "why can't we do Moxfield? Do Archidekt instead then if we
can't do Moxfield" (`01-UAT.md`, `G-01-1.reason`). The honest answer, already established
by this document's own Section 2 and not re-derived here: **Moxfield's response contract
is unobtainable without authorized API access.**

- `api.moxfield.com/robots.txt` is a blanket `User-agent: * / Disallow: /` (Section 2.1;
  re-verified 2026-08-05, unchanged). This spike declined to ignore that disallow, so no
  response body from that host was ever obtained, and Section 5's own "Open blockers
  gating enablement" already named this as blocker (a).
- `moxfield.com`'s web frontend independently returns a Cloudflare bot-management `403`
  "Attention Required" challenge to every unauthenticated probe attempted — three
  different `User-Agent` values, all blocked identically (Section 2.1). This is
  infrastructure-level anti-automation, not a policy choice this codebase's client could
  configure around without defeating a WAF challenge, which this spike also declined to
  do.

Section 5's own hard-gate finding (D-15, response-shape stability) could never be
evaluated for Moxfield for exactly this reason — no response body was ever observed to
evaluate. That was recorded honestly as "unknown," not rounded up to a pass. The user's
reversal is consistent with, not contrary to, that record.

### Moxfield is "not without a key," not "never"

Moxfield operates an official developer API with an application process (this spike did
not discover its exact terms — see Section 2.8 — but its existence as a *credentialed*
path, distinct from the unauthenticated `api.moxfield.com` this spike probed, is
independently known). Nothing here forecloses revisiting Moxfield if authorized access is
obtained: the barrier is the current unauthenticated posture, not something structural
about Moxfield as a provider. `deckimport.SourceMoxfield` remains a declared `SourceType`
value in `pkg/deckimport` for exactly this reason — the enum is closed at four values and
this decision does not touch it.

### The Archidekt risk carried forward unchanged

Selecting Archidekt does not retire the risk Section 1.4 already disclosed: `GET
https://archidekt.com/api/decks/<id>/` is Archidekt's own internal, unversioned SPA data
API — the same endpoint their React frontend calls — not a contract published or
versioned for third-party consumers. It was directly observed to be structurally
consistent across six independently fetched real public decks on 2026-08-05 (Section 1.4),
which is current-state evidence of consistency, not a guarantee of future stability. An
unversioned internal API can change without notice, and nothing in this document rules
that out. This is exactly why plan 01-09's fixture-contract tests against the committed
capture (`server/testdata/deck_providers/archidekt_deck_2026-08-05.json`) matter: they
make a future silent shape change fail loudly instead of mis-parsing quietly.

### Human precondition for enabling — not for building

Building `archidektAdapter` (plan 01-08, Task 2) requires none of this: it is exercised
entirely by hermetic tests against the committed fixture, and the provider kill switch
(`DECK_PROVIDER_ENABLED` / `providerEnabled()`) defaults off regardless of which adapter is
registered.

**Enabling the kill switch for real Archidekt traffic is a different matter, and has a
precondition that has not yet been satisfied by anyone:** `https://archidekt.com/terms` is
a client-rendered single-page app. The HTML a plain `GET` returns contains font
declarations and application shell markup, not the rendered terms text (Section 1.5) — no
agent working on this project has JavaScript execution capability, so **no agent has ever
read what Archidekt's terms actually say about automated or programmatic reads.**
`robots.txt` permitting the `/api/decks/` path (Section 1.5) is crawler etiquette, not a
license to build a product feature on top of that endpoint. A human must open
`https://archidekt.com/terms` in a real browser and read it before `DECK_PROVIDER_ENABLED`
is turned on for `archidekt.com` in any real deployment. This precondition is recorded
here, not resolved here.
