# vedh for iPhone

Native SwiftUI MVP for the existing vedh GraphQL service.

## Included in the MVP

- Sign up, sign in, sign out, and Keychain-backed session persistence
- List the signed-in player's games
- Create an EDH game with an optional commander and decklist
- Join from a game ID or `vedh.xyz/join/...` invite
- Poll current game state every three seconds
- Show turn, phase, priority, life totals, stack, and card-zone summaries
- Change your own life total
- Pass priority, advance phase, and claim a win when authorized
- Share an invite link

The MVP deliberately defers drag-and-drop card movement, draw/mill/scry controls,
WebSocket subscriptions, game history charts, and commander-damage tracking.

## Prerequisites

- Full Xcode with an iOS Simulator runtime (Command Line Tools alone are not enough)
- XcodeGen (`brew install xcodegen`)

## Generate and run

```sh
cd ios
xcodegen generate
open Vedh.xcodeproj
```

Select the `Vedh` scheme and an iPhone simulator, then run. The generated app
uses `https://api.vedh.xyz/graphql` by default. Change
`VEDH_GRAPHQL_ENDPOINT` in `project.yml` for a local or staging environment and
regenerate the project.

## Checks

The Foundation-only core can be compiled without Xcode:

```sh
swift build
swift run VedhCoreCheck
```

With Xcode and an iOS runtime installed:

```sh
xcodebuild \
  -project Vedh.xcodeproj \
  -scheme Vedh \
  -destination 'platform=iOS Simulator,name=iPhone 16 Pro' \
  test
```

See [the MVP plan](../docs/plans/2026-09-20-vedh-ios-mvp.md) for the acceptance
matrix and TestFlight gate.
