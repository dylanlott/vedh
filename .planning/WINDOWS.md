---
schema_version: 1
open_count: 1
waived_count: 0
fixed_count: 0
total_count: 1
last_updated: 2026-08-05T21:16:43.165Z
---

# Broken Windows Ledger

> Cross-phase defect register. `/gsd-ship` blocks while `open_count > 0`.
> Waive with `gsd-tools windows waive <id> "<reason>"` (reason required).
> Mark fixed with `gsd-tools windows fixed <id>`.

| id | phase | kind | file | line | description | status | reason | recorded_at | resolved_at |
|----|-------|------|------|------|-------------|--------|--------|-------------|-------------|
| 1 | 01 | stub | server/deck_providers.go | 395 | moxfieldAdapter.normalizeToDeckText always returns errMoxfieldContractUnverified -- no Moxfield field mapping is implemented because the response contract has never been observed (api.moxfield.com/robots.txt disallows automated access; no authorized sample response exists) | open |  | 2026-08-05T21:16:43.165Z |  |

````json
[
  {
    "id": 1,
    "kind": "stub",
    "phase": "01",
    "file": "server/deck_providers.go",
    "line": 395,
    "description": "moxfieldAdapter.normalizeToDeckText always returns errMoxfieldContractUnverified -- no Moxfield field mapping is implemented because the response contract has never been observed (api.moxfield.com/robots.txt disallows automated access; no authorized sample response exists)",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-08-05T21:16:43.165Z",
    "resolved_at": null
  }
]
````
