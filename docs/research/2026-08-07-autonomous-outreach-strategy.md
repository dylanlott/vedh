# Autonomous Outreach Strategy for vEDH — Research

**Date:** 2026-08-07
**Status:** Research, not yet a plan
**Relates to:** `docs/plans/2026-05-15-low-budget-launch-campaign.md` (manual predecessor),
`docs/product/2026-07-23-deck-to-game-activation-prd.md` (funnel + event vocabulary),
`docs/analytics/product-event-vocabulary.md` (attribution surface)

---

## 0. Method and limits of this research

What was directly read: Discord's Community Guidelines, Discord's Platform Manipulation
policy explainer, Discord Developer Policy summaries, and search-level summaries of the
EU AI Act Article 50 guidance, California SB 1001, Gmail/Yahoo bulk-sender requirements,
Google's scaled-content-abuse policy, and Anthropic's usage policy.

**What could not be verified from this environment:** every `reddit.com` and
`redditinc.com` URL is network-blocked here (WebFetch refused; `curl` against
`/about/rules.json` returned nothing parseable for r/EDH, r/magicTCG, r/CompetitiveEDH,
r/mtg). So the Reddit specifics below are stated from widely-reported secondary summaries
of Reddit's published self-promotion guidance, **not** from a first-hand read of the
current r/EDH sidebar. Before any Reddit activity, a human must read each target
subreddit's rules page directly. Treat this as an open precondition, the same way
`docs/research/deck-provider-feasibility.md` treats the unread Moxfield/Archidekt ToS.

Competitor names in §7 came from search results and are unverified beyond that.

---

## 1. The core finding

**The line between outreach and abuse is not "how much is automated." It is "which side
of the consent boundary the automated action lands on."**

Most writing about "autonomous outreach" gets this wrong by treating automation itself as
the variable to tune — send fewer cold DMs, personalize the template harder. That framing
produces exactly the 2026 outcome the market is now living with: a reported median
38-point sender-reputation drop within 90 days of scaling agentic outbound, 50–70% of
teams churning off AI SDR tools within three months, and buyers who have simply learned to
filter machine-written contact. Turning the volume down on an abusive channel does not
make it a good channel; it makes it a slow bad channel.

The variable that actually matters is consent. An action is safe to automate **without a
human in the loop** when the recipient has already granted permission for that class of
contact — they installed your bot, they clicked your invite link, they subscribed, they
came to your site. An action is **never** safe to fully automate when the recipient has
not — a Reddit comment, a cold DM, an unsolicited email, a post into someone else's
community.

This yields a four-tier architecture. The agent gets full autonomy in tiers 0–2 and is
restricted to *research, drafting, and queueing* in tier 3. That restriction is the whole
design; everything else is plumbing.

---

## 2. Constraint map

These are hard constraints, not preferences. They define what tier 3 means.

| Constraint | Source | What it forbids |
|---|---|---|
| No self-bots / user-bots | [Discord Community Guidelines #14](https://discord.com/guidelines) — "Do not use self-bots or user-bots. Each account must be associated with a human, not a bot." | Automating *your own* Discord account to post or DM. Full stop. Not a rate-limit question. |
| No unsolicited bulk messages | [Discord Community Guidelines #13](https://discord.com/guidelines) | Mass DM, and also *selling or building* the tooling for it. |
| Bot apps may not market to users | [Discord Developer Policy](https://support-dev.discord.com/hc/en-us/articles/8563934450327-Discord-Developer-Policy) | A vEDH bot may not DM users without explicit permission, and its messages must relate to the app's function — no unrelated marketing. |
| No inauthentic engagement | [Discord Community Guidelines #15](https://discord.com/guidelines), [Platform Manipulation policy](https://discord.com/safety/platform-manipulation-policy-explainer) | Fake members, join-for-join, engagement inflation. |
| Reddit self-promotion ratio | Reddit's published self-promotion guidance, ~90/10 participation-to-promotion, enforced per-account not per-post; individual subreddits enforce stricter ([summary](https://redship.io/blog/reddit-self-promotion-rules), [summary](https://indexly.ai/glossary/reddit-self-promotion-rules)) | An account that contributes primarily links to a project it benefits from is spam by definition — regardless of post quality. **Unverified against r/EDH's actual sidebar — see §0.** |
| GDPR: no B2C legitimate interest | [analysis](https://litemail.ai/blog/gdpr-legitimate-interest-cold-email-2026), [analysis](https://growleads.io/blog/is-cold-email-legal-gdpr-can-spam-2026/) | Cold email to personal addresses (players, not businesses) requires consent. There is no legitimate-interest route for B2C in most EU jurisdictions. vEDH's audience is consumers. **This kills cold email to players outright.** |
| Bulk sender requirements | [Google/Yahoo/Microsoft 2026 requirements](https://redsift.com/guides/bulk-email-sender-requirements) | Above 5k/day: SPF+DKIM+DMARC with alignment, RFC 8058 one-click unsubscribe, spam complaints under 0.30%. Since Nov 2025 non-compliant mail is rejected outright, not foldered. |
| EU AI Act Art. 50 | Applicable since **2 August 2026** — five days ago ([Commission guidance](https://digital-strategy.ec.europa.eu/en/policies/guidelines-transparency-ai-generated-content), [Art. 50](https://artificialintelligenceact.eu/article/50/)) | An AI system interacting directly with a person must be designed so the person knows it's a machine. Penalties up to €15M or 3% turnover. Applies to an EU-reachable support/outreach bot. |
| California SB 1001 (B.O.T. Act) | [Perkins Coie](https://perkinscoie.com/insights/update/i-am-robot-californias-new-law-requires-disclosure-use-bots) | Undisclosed bot communication to incentivize a sale. Note the real scope: it binds conduct *on platforms with 10M+ monthly US users* — so Reddit/Discord/X yes, vedh.xyz itself no. Disclose anyway; the cost is one line of copy. |
| Google scaled content abuse | [Search spam policies](https://developers.google.com/search/docs/essentials/spam-policies) | Mass-produced low-value pages. AI authorship is not the trigger; absence of value is. Rules out programmatic SEO pages generated per commander/card without genuine substance. |
| Anthropic usage policy | [AUP](https://support.claude.com/en/articles/9528712-exceptions-to-our-usage-policy) | Generating or facilitating spam, deceptive content, and coordinated inauthentic behavior. If the agent runs on Claude, tier-3 auto-send is a policy violation independent of any platform's rules. |
| `robots.txt` and ToS respect | This repo's own precedent — `docs/research/deck-provider-feasibility.md` | The Moxfield decision (blanket `Disallow: /` treated as dispositive, human ToS read as a precondition) is the standard. Outreach tooling inherits it: no scraping community member lists, no harvesting profile emails. |

Two consequences worth stating plainly, because they remove the tactics people usually
reach for first:

1. **Cold email to Commander players is not available.** They are consumers, on personal
   addresses, in a GDPR-covered population you cannot reliably geofence. Consent-first
   list building is the only email path.
2. **Automated Reddit and Discord participation is not available.** Not throttled —
   unavailable. Discord bans account automation outright; Reddit's ratio rule is
   account-level and a promotion-only account fails it by construction.

---

## 3. The recommended architecture: four consent tiers

### Tier 0 — The product's own invite loop (fully autonomous; consent is intrinsic)

**This is the highest-leverage autonomous outreach channel vEDH has, and it is already
specified in the activation PRD.**

Commander is a four-player format. Every hosted table is, structurally, an outreach event
in which a *friend* — not a marketer — delivers the message. The recipient has the
strongest possible consent signal: they were personally invited to a game they want to
play. The PRD's guest-first `/join/:id` flow removes the signup wall that was previously
destroying that loop's conversion.

The arithmetic: each activated host produces up to three invited views per table. At the
PRD's own targets (70% invite activation), one host who plays a single game yields ~2
activated players. No external channel available to a pre-launch indie tool competes with
that, and none of it requires contacting a stranger.

Autonomous work that belongs here — all of it safe, none of it needing a human gate:

- Rich OG/embed previews on `/join/:id` so a link pasted into a Discord DM renders as a
  real table card (format, players, capacity) rather than a bare URL. Link unfurling is
  the single cheapest conversion lever in an invite-driven product.
- Post-game recap artifacts worth sharing on their own merit, from data vEDH already has.
- Agent-run continuous monitoring of the funnel drop-off between `invite_viewed` →
  `join_started` → `player_joined` → `board_ready`, with the biggest weekly regression
  filed as a ticket.

**Recommendation: spend the first outreach budget here, not on any external channel.**
An autonomous system pointed at a leaky invite loop mostly automates the leak.

### Tier 1 — Owned surfaces (autonomous publish, human spot-check)

Surfaces vEDH controls, where the audience arrived voluntarily: the site, changelog, blog,
docs, GitHub releases, and vEDH's *own* social accounts (an official project account
posting about the project is not covered by the self-bot prohibition; automating a
*personal* account is a different thing and should be avoided).

The agent has real work here and it is genuinely autonomous:

- Read `git log` and the product-event tables; write a build-in-public update **only when
  something shipped**. Event-triggered, never calendar-triggered — a weekly cadence with
  nothing to say produces exactly the mass-produced filler Google's scaled-content-abuse
  policy targets and readers correctly ignore.
- Maintain a public changelog and RSS feed.
- Keep landing-page copy synchronized with what the product actually does (drift here is
  the most common cause of bad first impressions in fast-moving beta products).

Guardrail: a human reads the first ~10 published items before the review gate loosens.
Cap at one substantive post per week regardless of how much shipped.

### Tier 2 — Invited surfaces (autonomous, consent granted per install/subscribe)

Permission was explicitly granted by an act of installation or subscription. Automation
here is not merely tolerated — it is the expected behavior of the thing they installed.

- **A vEDH Discord app that server admins install.** `/vedh table` creates a table and
  posts the join link in-channel. The admin's install *is* the consent, which is what
  makes this categorically different from DMing that server's members. Playgroup discovery
  in Commander happens in Discord; this meets it where it is. Hard constraints from the
  Developer Policy: never DM a user unprompted, never send anything unrelated to the app's
  function, and label the bot as a bot.
- **Double opt-in email**, RFC 8058 one-click unsubscribe, SPF/DKIM/DMARC aligned,
  complaint rate watched against the 0.30% ceiling. Lifecycle mail to people who asked for
  it (your table is still open, your game recap is ready) is legitimate and automatable.
- **Support and feedback replies.** Someone who writes to you has invited a response.
  Support-driven growth is underrated and fully automatable up to the review gate.

### Tier 3 — Uninvited human surfaces (agent researches and drafts; **a human sends**)

Reddit, community Discords, cold DMs to creators, forum participation, Show HN.

The agent may: identify relevant threads and communities, read and summarize their posted
rules, draft a reply or post, and queue it with its rationale. The agent may **not** hold
credentials for these platforms, post, comment, vote, or DM.

A human reads, edits, and sends from their own account. This is not a compliance
formality — on Reddit the 90/10 ratio is satisfied by *genuine participation*, which is
precisely the part that cannot be delegated. An account that only ever emits queued
promotional drafts fails the rule no matter who clicks send.

Practical shape: the agent produces a short daily brief — 3–5 threads where vEDH is
genuinely relevant, each with the subreddit's rules quoted and a draft reply that answers
the person's actual question and mentions vEDH only if that mention is the honest answer.
Most days the correct output is "nothing relevant today," and the system must be able to
say that without penalty. A queue that always has five items is a queue that is
manufacturing relevance.

---

## 4. Concrete system design

Pipeline: **Sense → Draft → Gate → Publish → Measure**, run as a scheduled job.

```
  Sense    git log, product_events, support inbox, public mentions, provider health
    │
  Draft    per-surface artifact + rationale + the rules it must satisfy
    │
  Gate     tier 0-1: auto (spot-check)  │  tier 2: auto within consent  │  tier 3: human
    │
  Publish  owned surfaces + consented channels only
    │
  Measure  campaign source → existing product_event funnel
```

**Attribution is already built.** `landing_primary_cta` and `invite_viewed` both carry a
campaign source field in the PRD's event vocabulary, and the PRD already forbids storing
deck contents, URLs, credentials, or IPs in event payloads. The outreach system should
write into that existing vocabulary rather than inventing a parallel analytics path — it
inherits the privacy properties for free.

**Storage:** one `outreach_artifact` table — surface, tier, payload, rationale, state
(`drafted` / `approved` / `sent` / `rejected`), approver, timestamps. Every outbound
action, including autonomous tier 0–2 ones, gets a row. Without this you cannot answer
"what did the agent do last week," which is the question that matters the first time
something goes wrong.

### Guardrails

- **Suppression list.** Anyone who says stop, on any channel, is never contacted again on
  any channel. Checked before every send. Honor within 24–48h, not CAN-SPAM's 10 business
  days — GDPR expects prompt, and it costs nothing to just do it immediately.
- **Hard rate ceilings in code**, per surface, per day, counted in Postgres — not prompt
  instructions. A ceiling a model can talk itself past is not a ceiling.
- **Kill switch.** One env flag halts all outbound. Default-off in dev.
- **Dry-run default.** New surfaces start in draft-only mode and are promoted manually.
- **Disclosure.** Anywhere the agent itself speaks — the Discord app, any support
  autoresponder — it identifies as a bot. Required by AI Act Art. 50 for EU users;
  required by SB 1001 on large platforms; correct everywhere else regardless.
- **Abuse tripwire.** Track removals, reports, blocks, unsubscribes, and complaint rate
  per surface. Any surface crossing threshold auto-suspends and pages a human. This is the
  metric that distinguishes a system that is *actually* not abusive from one that merely
  intends not to be.
- **No identity fabrication, ever.** No sockpuppets, no vote manipulation, no purchased
  engagement, no bot posing as a person, no fake reviews. This is simultaneously banned by
  Discord, Reddit, the FTC, and the Anthropic AUP.
- **No harvesting.** No scraping member lists or profile emails from communities, no
  fetching anything behind a `robots.txt` disallow. Repo precedent already set.

---

## 5. Explicit anti-patterns

Things that will be suggested and should be refused:

| Anti-pattern | Why |
|---|---|
| Automating a personal Reddit/Discord account | Discord bans account automation outright; Reddit's ratio is account-level. Instant-loss move. |
| Cold email to scraped player addresses | B2C + GDPR = no legitimate-interest route. Also destroys domain reputation. |
| Mass-DMing Discord server members | Guideline #13 and #14 simultaneously. |
| Programmatic SEO pages per commander/card | Scaled content abuse; March 2026 update specifically targeted this. |
| "Helpful" bot replies to MTG mentions across Reddit/X | This is the exact pattern that discredited AI SDRs. Uninvited by definition. |
| Paying for or trading engagement | Inauthentic engagement; banned everywhere. |
| Posting to 10 subreddits at launch | Each has its own rules; cross-posting promotion is the canonical spam signature. |

---

## 6. First 30 days

1. **Week 1 — Tier 0.** OG/embed previews for `/join/:id`. Instrument invite-loop
   drop-off against the existing event vocabulary. Nothing external.
2. **Week 2 — Tier 1.** Changelog + RSS; agent drafts build-in-public updates from real
   commits, human reviews all of them.
3. **Week 3 — Tier 3 in draft-only.** Agent produces the daily relevance brief. Human
   reads target subreddit rules *first-hand* (unblocking §0), participates genuinely, and
   sends nothing promotional for at least two weeks.
4. **Week 4 — Tier 2 groundwork.** Scope the Discord app. Stand up double opt-in email
   with authentication and one-click unsubscribe before collecting a single address.

Success is measured on the PRD's existing funnel — activated players and returning pods —
not on volume of outreach emitted. A system that sent nothing this week because nothing
was worth sending is functioning correctly.

---

## 7. Positioning note (unverified, worth checking)

Search surfaced a more crowded 2026 field than the launch-campaign doc assumes: Playgroup
Live (free browser tabletop, 2–6 players, ~3,030 pickup games logged in July 2026),
Convoke (webcam, positioned as the direct SpellTable replacement), TableCommander
(**account-free guest play** — the same wedge the activation PRD is building), plus EDHplay
and EDHLAB. PlayEDH runs playgroup discovery through Discord with deck verification.

Two implications: guest-first is table stakes rather than a differentiator, and Discord is
demonstrably where playgroup formation happens — which strengthens the tier-2 Discord app
over any tier-3 broadcast tactic. Worth a human verification pass before positioning copy
is written.

---

## 8. Open decisions

1. Is the goal players, or creators/community operators? Creators are a legitimate B2B-ish
   cold-outreach population (GDPR legitimate interest is arguable, volumes are tiny,
   personalization is real). Players are not. This changes tier 3 substantially.
2. Does vEDH have a sending domain and list infrastructure, or does tier 2 email start
   from zero? (Affects week-4 scope.)
3. Is a Discord app in scope? It is the strongest tier-2 surface and a real engineering
   project, not a marketing task.
4. Who is the human approver for tier 3, and what is their realistic daily throughput?
   The gate only works if someone actually staffs it.

---

## Sources

- [Discord Community Guidelines](https://discord.com/guidelines)
- [Discord Platform Manipulation Policy Explainer](https://discord.com/safety/platform-manipulation-policy-explainer)
- [Discord Developer Policy](https://support-dev.discord.com/hc/en-us/articles/8563934450327-Discord-Developer-Policy)
- [EU AI Act Article 50](https://artificialintelligenceact.eu/article/50/) · [Commission transparency guidance](https://digital-strategy.ec.europa.eu/en/policies/guidelines-transparency-ai-generated-content) · [Sidley: preparing for 2 Aug 2026](https://datamatters.sidley.com/2026/06/24/eu-ai-act-transparency-obligations-preparing-for-compliance-by-2-august-2026/)
- [Perkins Coie on California SB 1001](https://perkinscoie.com/insights/update/i-am-robot-californias-new-law-requires-disclosure-use-bots) · [SB 1001 text](https://leginfo.legislature.ca.gov/faces/billTextClient.xhtml?bill_id=201720180SB1001)
- [Google Search spam policies](https://developers.google.com/search/docs/essentials/spam-policies)
- [Bulk email sender requirements 2026](https://redsift.com/guides/bulk-email-sender-requirements)
- [GDPR legitimate interest and cold email](https://litemail.ai/blog/gdpr-legitimate-interest-cold-email-2026) · [Is cold email legal in 2026](https://growleads.io/blog/is-cold-email-legal-gdpr-can-spam-2026/)
- [Anthropic usage policy](https://support.claude.com/en/articles/9528712-exceptions-to-our-usage-policy)
- [The case against AI SDRs](https://www.digitalapplied.com/blog/case-against-ai-sdrs-contrarian-analysis-2026) · [Why first-gen AI SDRs failed](https://www.kwanzoo.com/blog/why-first-gen-ai-sdrs-failed) · [AI SDR outbound results 2026](https://www.harborbd.com/blogs/ai-sdr-outbound-results-2026)
- [Reddit self-promotion rules summary](https://redship.io/blog/reddit-self-promotion-rules) · [90/10 rule](https://indexly.ai/glossary/reddit-self-promotion-rules) *(secondary sources; reddit.com unreachable from this environment)*
- [SpellTable alternatives 2026](https://playgroup.gg/playgroup-live/spelltable-alternative) · [How to play Commander online](https://draftsim.com/how-to-play-commander-online/)
