# vEDH TikTok Marketing Strategy

**Status:** working strategy  
**Scope:** vEDH, the multiplayer TCG/Commander board-state tracker at `vedh.xyz`  
**Horizon:** 90 days  
**Operating model:** automate preparation and measurement; keep taste, rules accuracy, community interaction, and publishing under human control.

The authoritative cross-channel targets and definitions live in [social-engagement-goals.md](social-engagement-goals.md). This document defines TikTok execution within that scorecard.

## Recommendation

Lead with one sharp promise:

> **Commander table chaos, made legible.**

Do not market vEDH as “a platform for any TCG” yet. That is technically broad and commercially vague. TikTok should acquire one community first: Commander players who already play remotely, use a second screen, organize a regular pod, or care about complicated board states.

The content should look like a builder showing useful software to other players—not a SaaS brand performing enthusiasm. Screen captures, real play situations, dry observations, and visible product behavior are the raw material.

## Objective

The first 90 days are a validation program, not a follower campaign.

**North-star metric:** weekly activated pods.

An **activated pod** is a new table where:

1. the host arrives from TikTok or a TikTok-assisted referral;
2. at least one other player joins; and
3. the group records at least 10 game actions within 24 hours.

This is stronger than signups: vEDH becomes useful when a group uses it together.

### Provisional 30-day targets

Reset these after the first 20 posts establish a baseline.

- 20 published posts
- 100 qualified visits from TikTok
- 20 completed signups
- 5 activated pods
- at least 2 repeatable formats that beat the rolling 10-post median

Follower count is diagnostic, not a goal.

## Audience

### Primary

**Commander pod organizers and remote/hybrid players**

- play weekly or more
- manage four-player board states
- already tolerate a laptop, webcam, or second screen at the table
- feel friction around priority, counters, hidden zones, turn flow, or reconstructing what happened
- can bring three additional users if the host experience works

### Secondary

**Commander creators, judges, and rules-heavy players**

- need clear visual examples
- can demonstrate complicated states better than a generic product video
- generate qualified comments and product feedback

### Later

Other TCG communities. Pokémon support and “any TCG” messaging should become separate experiments only after Commander activation is repeatable.

## Positioning

### One-line description

> vEDH is a shared live board state for Commander: zones, life, priority, and game history stay synced.

### Message hierarchy

1. **Pain:** four-player state gets hard to read.
2. **Proof:** show one action updating for everyone.
3. **Payoff:** fewer state disputes; a cleaner record of the game.
4. **CTA:** start a table at `vedh.xyz`.

### Voice

- lowercase where practical
- direct and technically accurate
- dry, not cynical
- builders talking to players
- no “revolutionize your game night,” “seamless platform,” or generic founder inspiration

Example:

> 47 permanents. 12 triggers. nobody remembers who has priority.  
> we gave the table a shared state.

## Account and publishing setup

Use a **TikTok Business Account**. The immediate website link is worth more than unrestricted trend audio for a product acquisition channel. TikTok says Business Accounts can add a website link; Personal Accounts need at least 1,000 followers. Business Accounts use TikTok’s Commercial Music Library, so build the creative system around founder voiceover, interface sounds, and cleared background audio. See TikTok’s [account-type guidance](https://support.tiktok.com/en/using-tiktok/growing-your-audience/switching-to-a-creator-or-business-account?lang=en) and [Commercial Music Library instructions](https://ads.tiktok.com/resources/help/article/how-to-use-the-commercial-music-library?lang=en-GB).

Profile:

- **name:** vEDH
- **bio:** `shared board state for commander. zones, priority, game logs.`
- **link:** a first-party tracked route such as `https://vedh.xyz/tiktok`
- **pinned posts:** what it is; 20-second demo; founder/build story

Do not build a custom posting integration during phase one. TikTok supports direct and draft uploads through its Content Posting API, but public Direct Post requires an approved `video.publish` scope and an audited client; unaudited clients are private-only. A native or approved scheduling flow is lower-risk. See the official [Content Posting API guide](https://developers.tiktok.com/docs/en/content-posting-api-get-started) and [product overview](https://developers.tiktok.com/products/content-posting-api).

## Content system

### Five pillars

| Pillar | Job | Examples | Share of posts |
|---|---|---|---:|
| table chaos | earn attention through a recognized Commander problem | triggers, priority, counters, unreadable remote boards | 30% |
| product proof | show the interface doing one useful thing | move a card, pass priority, resolve the stack, review a log | 30% |
| post-game data | make the game record interesting | life chart, turning point, event volume, “prove what happened” | 15% |
| build notes | make the builder credible and the product human | why a control exists, bug fixed from a pod’s feedback | 15% |
| comment replies | convert questions and objections into native content | “does it support partner commanders?” | 10% |

### Repeatable formats

1. **Pain → proof (15–22s)**  
   recognizable table problem for 2–3 seconds; one product action; result; CTA.

2. **Before/after (12–18s)**  
   messy physical/remote state, then the synced vEDH view.

3. **One-feature walkthrough (25–40s)**  
   founder voiceover, one feature only, no feature inventory.

4. **Reply to comment (10–25s)**  
   display the real question; answer with the product or an honest limitation.

5. **Build log (20–35s)**  
   “a pod asked for X. here is what changed.”

### Creative rules

- show a face or human voice at least once each week
- put the product on screen by second 3 in proof posts
- use burned-in captions and large interface crops
- shoot/edit vertical at 1080×1920
- one idea per video
- show the result before explaining the implementation
- use real card-game language; manually verify rules claims
- avoid generic AI avatars and synthetic founder voices
- if realistic video or audio is materially AI-generated, apply TikTok’s required disclosure; TikTok says the label does not reduce distribution when the content otherwise complies. See its [AI-generated content guidance](https://support.tiktok.com/en/using-tiktok/creating-videos/ai-generated-content?authuser=0).

## Semi-automated production loop

### Weekly batch: 90–120 minutes

**1. Source capture — human**

- record 20–30 minutes of real or staged vEDH play
- capture a clean desktop/mobile viewport and, when useful, a physical table angle
- mark 5–10 moments: confusion, state change, payoff, failure, surprising data
- collect the week’s substantive comments and support questions

**2. Ideation — automated draft, human selection**

For each source moment, generate:

- three hooks
- one 15–20 second script
- one 25–35 second script
- three caption variants
- a suggested pillar and CTA

Human selects five concepts based on clarity and actual product truth. TikTok’s free Creative Center can be used once weekly to inspect current hashtags, videos, and patterns by region; it is an input, not a reason to force irrelevant trends. See [Creative Center](https://ads.tiktok.com/resources/help/article/creative-center) and its [Trends workflow](https://ads.tiktok.com/resources/help/article/how-to-use-trends?lang=en&redirected=2).

**3. Assembly — mostly automated**

- drop clips into one of five edit templates
- auto-transcribe and burn captions
- crop/zoom the product area
- normalize voice audio
- export a clean master with no platform watermark
- create caption, cover text, UTM slug, and alt-text draft

**4. Approval — human gate**

Before anything enters the publishing queue:

- product claim is visible and true
- card/rules statement is correct
- no private player information appears
- card art and music usage are appropriate
- AI labeling requirement is checked
- the first frame makes sense without audio
- CTA matches the linked destination

**5. Publish — human-assisted**

- queue five approved masters in TikTok’s native or an approved scheduler
- select cleared audio in the final TikTok flow
- do a final preview for caption placement and interface legibility
- publish at consistent test windows; do not overfit posting time before there is account data

**6. Community — human**

- spend 15–20 minutes after publishing answering real questions
- reply within 24 hours to every substantive early comment
- turn the best question into a reply video within 72 hours
- never automate generic comments, DMs, follows, or engagement

**7. Reporting — automated**

At 24 hours and 7 days, write post metrics into the tracker. Generate a weekly report with:

- rolling median by format and pillar
- attributed visits, signups, and activated pods
- top three hooks
- comments that reveal objections or feature demand
- “remake / iterate / retire” recommendation for every concept

## Cadence

Publish **five times per week** for the first eight weeks:

- Monday: table-chaos observation
- Tuesday: product proof
- Wednesday: build note or founder explanation
- Thursday: product proof or comment reply
- Friday: post-game data or pod prompt

Weeks 9–12: maintain five posts only if batching stays under two hours. Otherwise publish the best three. Consistency is useful; exhausted filler is not.

## Starter hooks

1. `47 permanents. 12 triggers. one player says “whose turn is it?”`
2. `we gave a commander game a transaction log.`
3. `the stack is not a suggestion.`
4. `your webcam shows the cards. this shows the game state.`
5. `this is what priority looks like when the whole pod can see it.`
6. `someone gained 18 life. someone else disagreed. the log did not.`
7. `a commander game ended. now prove what happened.`
8. `we made hidden zones hidden and everything else obvious.`
9. `the feature nobody asks for until turn nine.`
10. `building commander software means modeling arguments as state.`

## Conversion path

Build a dedicated TikTok landing path before scaling traffic.

**Required page:** `vedh.xyz/tiktok`

Above the fold:

1. Commander-specific headline
2. silent 10–15 second product loop
3. `start a table` CTA
4. three proof points: shared zones, priority/turn flow, game history
5. no umbrella-brand explanation

Preserve UTMs through signup:

```text
utm_source=tiktok
utm_medium=organic
utm_campaign=commander_launch
utm_content=<post-slug>
```

Track:

- `landing_view`
- `signup_started`
- `signup_completed`
- `table_created`
- `invite_link_copied`
- `player_joined`
- `game_activated`
- `game_completed`

The current backend exposes aggregate signup/activity metrics, but the repository does not show TikTok attribution through signup. Add attribution before paid spend.

## Measurement and decisions

### Post-level metrics

- average watch time
- full-watch percentage
- shares and saves
- substantive comments
- profile views
- profile-to-link click rate
- attributed visits and signups
- activated pods

### Decision rules

- **Remake:** a post generates at least 2× the rolling 10-post median for qualified visits, activated pods, shares, or profile views. Reuse the concept with a new opening within 72 hours.
- **Iterate:** strong watch/share signal but weak clicks. Tighten the product reveal and CTA.
- **Fix funnel:** strong clicks but weak signup or activation. The content is working; the landing/onboarding is not.
- **Retire:** a format runs three times below median with no unusually useful comments or conversion signal.

Do not judge a concept from one post. Judge a format after three executions.

## Creator and paid layer

Start in weeks 5–8, after the account has at least two organic formats that work.

### Creator program

- identify 10–15 Commander creators with roughly 5k–50k followers and active comments
- prioritize remote-play, judge/rules, deck-tech, and regular-pod creators
- offer a short co-design session and a private table for their pod
- ask for honest use, not a scripted endorsement
- license strong creator clips explicitly before reusing them
- measure activated pods per creator, not impressions alone

### Paid tests

Do not run cold ads against unproven creative. When an organic post beats the rolling median and drives qualified visits:

- run a small 3–5 day Promote/Spark test
- optimize for site visits initially, then activated events when enough conversion data exists
- compare cost per activated pod, not cost per follower
- stop any paid test that produces clicks without group activation

## 90-day sequence

### Weeks 0–1: foundation

- create Business Account and profile
- build tracked TikTok landing route
- implement attribution and activation events
- produce five edit templates
- capture one staged four-player game
- prepare the first 10 posts

### Weeks 2–4: find the wedge

- publish five times weekly
- test all five pillars and at least three opening styles
- answer comments manually
- remake winners within 72 hours
- hold a weekly 30-minute metric review

### Weeks 5–8: concentrate

- allocate 70% of posts to the two best formats
- begin creator co-design outreach
- test one creator/pod collaboration per week
- improve onboarding at the highest-dropoff step
- run paid only on proven organic posts

### Weeks 9–12: prove repeatability

- document the content-to-activated-pod funnel
- create a referral moment after `invite_link_copied` or `game_completed`
- decide whether another TCG deserves a separate account experiment
- keep, change, or stop TikTok based on activated pods and retention—not reach alone

## Minimum viable operating stack

No custom social platform is required.

- **capture:** OBS or native screen recording
- **source library:** non-git folder with `inbox / selected / published`
- **editing:** five reusable vertical templates
- **queue:** TikTok native scheduler or approved scheduler
- **content tracker:** the accompanying CSV
- **metrics:** TikTok analytics plus first-party UTM/event data
- **weekly synthesis:** generated from the tracker, reviewed by a human

## First actions

1. Add TikTok attribution and define `game_activated`.
2. Create `vedh.xyz/tiktok` with Commander-specific copy.
3. Record one staged game and extract 10 moments.
4. Produce the first five posts from the launch calendar.
5. Publish for four weeks before expanding scope or building posting infrastructure.
