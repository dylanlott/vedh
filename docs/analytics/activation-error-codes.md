# Activation Error-Code Vocabulary

This is the documented form of the closed, server-owned vocabulary in
`server/activation_errors.go`. GraphQL activation failures carry the code in
`errors[].extensions.code`; the message contains product language only. Raw SQL,
driver, provider, and Go wrapper text stays in the server-side wrapped cause.

## The four codes

| Code | Resolver paths | Locked product copy | Required affordance |
|---|---|---|---|
| `provider_unavailable` | `previewDeck` when a configured provider fetch cannot complete | "We can't reach that deck link right now. Paste your decklist instead — it takes about 10 seconds." | **Paste instead** |
| `preview_error` | `previewDeck` when a fetched deck cannot be normalized into a preview | "We couldn't fully read a few cards. Fix the highlighted entries below, then continue — your original paste is untouched." | **Fix these cards** inline |
| `guest_session_error` | `guestSession`, `refreshGuestSession`, and guest account-claim flow failures | "We couldn't set up your seat at the table. Nothing you entered was lost. Try again in a moment." | **Try again** |
| `create_error` | Quick-start `createGame` transport and flow failures (wired by Plan 02-04) | "We couldn't create your table. Your deck and commander picks are still here — try again." | **Try again** |

## Closure and widening

The set is closed. An activation resolver must use one of the four dedicated
constructors; it may not forward an arbitrary code or underlying error string.
Widening the vocabulary requires updating the allowlist and closure test in
`server/activation_errors.go` / `server/activation_errors_test.go`, this document,
and the client's exhaustive copy-and-affordance map.
