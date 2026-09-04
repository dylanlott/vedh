# Landing Page WebGL Polish — Concept Backlog

**Status:** Concept #1 is planned (see `2026-08-25-hero-holofoil-implementation-plan.md`). #2–#6 are unstarted.

**Goal:** Give `LandingView.vue` a distinctive, physical-feeling motion layer that reads as "trading cards" rather than as generic startup-landing-page particle soup — without regressing bundle size, accessibility, or the no-JavaScript/no-WebGL fallback.

**Design constraints these concepts were shaped around:**

- The hero art is currently pure CSS: `.tcg-card-frame` plus `.hero-orbit` rings and `.hero-sigil` diamonds. Anything WebGL is **additive** — the CSS art stays as the fallback and must still look finished on its own.
- Palette is fixed by `src/styles/main.scss` tokens: warm brown `#3a3431`, ember `#f2a041`, flare `#f00876`, parchment `#fff4ed`. No effect introduces a hue outside that ramp — that discipline is what makes motion read as vEDH branding instead of stock shader art.
- Landing copy promises **live zones, shared counters, turn flow**. Effects that visualize those claims are worth more than effects that are merely pretty.
- The landing copy explicitly disclaims "borrowed game framing." No mana symbols, no five-colour pie, no WotC-adjacent iconography.

---

## The six concepts, ranked by payoff-per-effort

### 1. Holofoil pass over the existing hero card — PLANNED

Keep `.tcg-card-frame` as real DOM (real text, real a11y, real SEO). Overlay a WebGL canvas *inside* the frame — inheriting its `border-radius` and `overflow: hidden` for free — rendering only a holographic sheen: fresnel rim light plus a noise-warped diagonal diffraction band whose hue sweeps a constrained ember→flare→parchment ramp. Composite with `mix-blend-mode: screen` so the effect can only ever brighten, never wreck, the existing design. Pointer position drives both the shader's view vector and a damped CSS `rotate3d` tilt on a wrapper, so sheen and tilt share one spring.

- **Why it wins:** every TCG player has tilted a foil under a light. Instantly legible, costs one fullscreen quad, and discards none of the existing markup.
- **Cost:** one shader, no geometry.
- **Full plan:** `docs/plans/2026-08-25-hero-holofoil-implementation-plan.md`

### 2. Priority ring — replace `.hero-orbit` with a real one

The two CSS circles become a single tilted 3D ring carrying four player nodes. A light pulse travels node → node → node; on arrival the node blooms and fires a thin arc to the card at centre. This is *literally* "pass priority, log moves, stay synced" rendered as an object rather than asserted in body copy.

- **Build:** tube or `Line2` geometry with a gradient scrolling along UV, plus four additive sprites. A few hundred verts, effectively one draw call.
- **Animation state:** a single `t` value modulo 4. Cheap to author, cheap to run.
- **Pairs with #1** without competing — the ring moves slowly and continuously, the foil reacts fast to input. Different frequency bands.
- **Cost:** low. Best follow-up, because it *replaces* existing markup instead of adding a new surface.

### 3. The Stack, resolving

For the "Match flow tools" feature card: a vertical stack of translucent cards. Cards fly in from off-screen, land on top, then the top card dissolves upward in a spark burst and the next resolves. The LIFO stack is one of the few genuinely three-dimensional concepts in the game and nearly nobody visualizes it.

- **Build:** `InstancedMesh`, hard cap ~12 cards.
- **Cost:** medium — the difficulty is choreography and timing taste, not technique.

### 4. Scroll-scrubbed deck: fan → riffle → board

Tie the `.playbook` steps (`01 open a table / 02 sync your decks / 03 play faster`) to one continuous deck animation **scrubbed by scroll position** rather than played on a timer. 01 fans the deck out, 02 riffle-shuffles it, 03 snaps the cards into a board layout.

Scroll-scrubbing is the specific thing that separates "this page has an animation" from "this page feels authored" — the reader is driving it, so it never plays at the wrong moment.

- **Build:** `InstancedMesh` of 40–60 cards; per-instance matrices lerped between three authored keyframe poses. GPU cost is negligible; all the work is in the pose curves.
- **Cost:** highest of the six, and the biggest payoff. Deliberately queued *after* #1/#2 so the renderer plumbing already exists.

### 5. Living table felt (full-bleed background)

Replace the `.landing::before` / `.landing::after` radial gradients with a single fullscreen shader quad: warm felt grain, a slow-drifting hex weave, and concentric ripples emanating from the pointer like a life-total change crossing the table. One draw call, no geometry, and it upgrades *every* section rather than only the hero.

- **The discipline here is restraint.** If it is consciously noticeable while reading the copy, it is at least twice as strong as it should be.
- **Cost:** lowest of the six. The right pick for a fast standalone win.

### 6. Counter burst on the CTA

Hovering "Create a table" emits a handful of small bevelled counter discs that arc outward and fade. Pure micro-interaction, reuses the shared renderer, roughly thirty lines.

- **Cost:** trivial. Good filler work once the plumbing exists.

---

## Considered and rejected

- **Energy conduits between the `.feature-grid` cards.** Thin traced lines in 3D behind the DOM as cards enter the viewport. Sounds good, reliably looks messy over live text at arbitrary viewport widths, and there is no framing that fixes it.
- **Ambient drifting card motes.** The generic option. Cheap and low-risk, but adds no meaning, and #5 occupies the same "ambient layer" slot with far more character.
- **Route-transition card flip on "Start a table".** A screen-filling card flip wiping into `/signup`. Genuinely nice, but it delays the single most important conversion action on the page. Motion in front of a CTA has to earn its milliseconds and this does not.

---

## Shared plumbing (matters more than any individual effect)

Decisions that should hold across every concept above, so #2–#6 inherit them rather than relitigating them:

- **One canvas, one renderer** for the whole page, using `setScissorTest` plus per-section viewports (three's "many elements, one canvas" pattern). Browsers reap WebGL contexts around ~16 and you get intermittent blanking that is miserable to debug.
- **Everything is additive.** The CSS art is the fallback. Feature-detect; never gate content or layout on WebGL succeeding.
- **Lazy-load inside `onMounted`** via `await import('three')` behind an `IntersectionObserver`. `LandingView` is already a lazy route chunk, so three lands in its own chunk and never reaches `/play`, `/login`, or the board.
- **Skip `EffectComposer` entirely.** Postprocessing roughly doubles the cost for effects this small. Fake bloom by baking glow into additive sprites.
- **`prefers-reduced-motion: reduce`** renders exactly one frame and stops the loop — keeping the look while dropping the motion. Note that `LandingView`'s existing `fadeUp` and `floatCard` keyframes currently ignore this preference; whichever concept lands first should fix that.
- **Pause the RAF loop** on `document.hidden` and on `IntersectionObserver` exit. Clamp `setPixelRatio` to `Math.min(devicePixelRatio, 2)`. Cap ambient layers at 30fps.
- **Do not request `powerPreference: 'high-performance'`.** It forces the discrete GPU on laptops, which is an indefensible battery cost for a marketing page.
- **Below 720px, fall back to CSS.** The existing `@media (max-width: 720px)` block already shrinks the hero art, so there is a clean seam to fall back to.
- **Honour `navigator.connection.saveData`** by skipping the WebGL chunk entirely.

## Why three.js at all

For concept #1 in isolation, three.js is not required — a single fullscreen quad is about eighty lines of raw WebGL2 with zero dependency weight. The dependency is justified by #2, #3, and #4, which all need a scene graph, instancing, and camera handling. Concept #1 is therefore treated as the foundation investment: it pays for the renderer plumbing that the more ambitious concepts then reuse.

If #2–#4 are ever formally dropped, revisit this and strip back to raw WebGL.

## Suggested sequencing

1. **#1 Holofoil** — establishes plumbing, lifecycle, reduced-motion policy, and the capability gate.
2. **#5 Living felt** — reuses all of the above for a whole-page lift at near-zero marginal cost.
3. **#2 Priority ring** — replaces existing CSS markup, so it adds surface without adding clutter.
4. **#4 Scroll-scrubbed deck** — the showpiece, once the foundation is proven in production.
5. **#3 The Stack** and **#6 Counter burst** — opportunistic.
