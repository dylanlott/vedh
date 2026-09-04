# Hero Holofoil Implementation Plan

**Concept:** #1 from `docs/plans/2026-08-25-landing-webgl-polish-concepts.md`.

**Goal:** Make the landing page hero card feel like a physical foil card. The card tilts toward the pointer and a holographic sheen sweeps across it as the view angle changes. The existing DOM card is preserved exactly as-is; the effect is a brighten-only overlay that can be removed at any time with no visual debt.

**Architecture:** The WebGL canvas mounts *inside* `.tcg-card-frame`, so it inherits the frame's `border-radius` and `overflow: hidden` and needs no rounded-rect masking or bounding-rect synchronisation. It renders a single fullscreen quad with an orthographic setup — there is no 3D geometry at all, and the "3D" is entirely faked in the fragment shader from a pointer-derived view vector. Compositing is `mix-blend-mode: screen`, which is *only* capable of brightening; a shader bug can wash the card out but can never destroy the underlying design. Card copy stays real DOM text throughout, so accessibility, selection, and SEO are untouched.

Tilt is applied by CSS on a new wrapper element rather than in WebGL, so the DOM card and the canvas rotate together as one object for free. A single damped-spring value feeds both the CSS transform and the shader uniform, which is what keeps sheen and geometry from visibly desynchronising.

**Tech Stack:** Vue 3 (`<script setup lang="ts">`), Vite 5, SCSS, three.js `^0.185.1` with `@types/three` `^0.185.4`, Vitest + jsdom, Playwright.

**Non-goals for this plan:** the shared multi-section renderer (concepts #2–#6 introduce that), `deviceorientation` tilt, and any change to landing copy or layout.

---

### Task 1: Add the three.js dependency and prove the build gate

**Files:**
- Modify: `app/package.json`
- Inspect: `app/vite.config.ts`, `app/README.md`

**Step 1: Install**

```sh
cd app
npm install three@^0.185.1
npm install -D @types/three@^0.185.4
```

three.js does not ship its own type declarations and does not expose a `types` entry in `exports`; `@types/three` is required because `npm run build` runs `vue-tsc --noEmit` before Vite. Keep the two versions' minor numbers aligned — `@types/three` tracks three's release line and drifts quickly.

**Step 2: Confirm the toolchain actually accepts it**

```sh
npm run type-check
npm run build
```

`app/README.md` claims the frontend "uses Node 16 and won't build with any other version." That claim is already stale — the working tree builds on Node 24 with Vite 5 — but if any CI job really does pin Node 16, verify there before assuming three 0.185 installs. If Node 16 turns out to be a hard constraint, pin to an older three release line rather than fighting it, and record that decision here.

**Step 3: Confirm chunk isolation**

Inspect `dist/assets/` after the build. three must appear only in a chunk reachable from the `LandingView` chunk — never in the `index` entry chunk. If it leaks into the entry chunk, the async component boundary in Task 5 is wired wrong.

**Verify:** `npm run build` passes and no non-landing chunk grows.

---

### Task 2: Pure helpers — tilt smoothing and rect-to-UV

**Files:**
- Create: `app/src/composables/usePointerTilt.ts`
- Create: `app/src/composables/rectToUv.ts`
- Create: `app/__tests__/usePointerTilt.spec.ts`
- Create: `app/__tests__/rectToUv.spec.ts`

`app/src/composables/` is a new directory. It is the idiomatic Vue 3 location for reactive view helpers, and these do not belong in `services/` (which holds app-level concerns like Apollo and Scryfall) or `utils/` (pure domain helpers like `stack.ts`).

**Step 1: Frame-rate-independent smoothing**

The whole feel of this effect lives in one function. Do not use a naive `current += (target - current) * 0.1` — that couples feel to frame rate and behaves differently on a 144Hz display than on a throttled background tab.

Target shape:

```ts
/** Exponential smoothing that is invariant to frame duration. */
export function approach(current: number, target: number, rate: number, dt: number): number {
  return current + (target - current) * (1 - Math.exp(-rate * dt));
}
```

A `rate` of roughly `6` reads as responsive-but-weighted. Expose it as a constant so it can be tuned without hunting.

**Step 2: The composable**

`usePointerTilt(hostRef)` returns `{ tilt, pointer, isIdle }` where `tilt` is a `-1..1` pair derived from pointer offset relative to the host's centre (clamped, not normalised by distance — clamping keeps the effect stable when the pointer is far away) and `pointer` is `0..1` in card-local space for the specular hotspot.

Requirements:
- Listen on `window` for `pointermove`, not on the card. The tilt should begin responding before the pointer arrives, which is what makes it feel anticipatory rather than reactive.
- After ~2s without pointer movement, lerp the target toward a slow Lissajous drift so the foil keeps breathing on an idle page. This matters more than it sounds: a static foil looks broken.
- On touch devices, drive tilt from the idle drift only, plus `touchmove` while actively dragging. **Do not** call `DeviceOrientationEvent.requestPermission()` — a permission prompt on a marketing page is hostile and will cost more conversions than the effect gains.
- Register listeners with `{ passive: true }`. Remove every listener in `onUnmounted`.
- Read `matchMedia('(prefers-reduced-motion: reduce)')` and, when set, freeze the tilt at a fixed flattering angle (roughly `(0.25, -0.15)`) and skip all listener registration. Subscribe to that query's `change` event so a mid-session preference change is honoured.

**Step 3: Rect-to-UV**

The shader needs to know where the artboard and the text blocks are, in UV space, so it can put the sheen where it belongs. Measure rather than hardcode — the card is responsive and the `@media (max-width: 720px)` block changes `.tcg-card-artboard` height.

```ts
/** DOM rect -> WebGL UV rect (x, y, w, h). Flips the Y axis. */
export function rectToUv(child: DOMRect, host: DOMRect): [number, number, number, number] {
  const x = (child.left - host.left) / host.width;
  const w = child.width / host.width;
  const y = 1 - (child.bottom - host.top) / host.height;
  const h = child.height / host.height;
  return [x, y, w, h];
}
```

The Y flip is the easy thing to get wrong here and it fails *subtly* — the sheen lands on the rulebox instead of the artboard and looks merely "a bit off" rather than obviously broken. Test it explicitly.

**Verify:** `npm test`. Both helpers are pure and deterministic given `dt`, so they get real assertions rather than smoke tests: `approach` converges monotonically, produces identical results for one 32ms step versus two 16ms steps to within tolerance, and never overshoots. `rectToUv` maps a known child rect to known UV values including the flip.

---

### Task 3: Restructure the hero card DOM for tilt — CSS only, no WebGL

**Files:**
- Modify: `app/src/views/LandingView.vue`

This task ships a complete, independently valuable improvement with zero dependency on three.js. Do it first and confirm it in a browser before any WebGL exists.

**Step 1: Add a tilt wrapper**

`.tcg-card-frame` already owns `animation: floatCard 7s ease-in-out infinite`, and those keyframes animate `transform`. A JS-driven tilt on the same element would fight the keyframes and lose. Wrap instead, so the two transforms compose across two elements:

```html
<div class="tcg-card-tilt" :style="tiltStyle">
  <div class="tcg-card-frame">
    <!-- unchanged -->
  </div>
</div>
```

```scss
.tcg-card-tilt {
  transform: perspective(1200px)
             rotateX(var(--tilt-x, 0deg))
             rotateY(var(--tilt-y, 0deg));
  transform-style: preserve-3d;
  will-change: transform;
}
```

Drive it by writing CSS custom properties from the composable, not by rebuilding a transform string in JS each frame — custom properties keep the style object stable and the diff cheap.

Cap the rotation at roughly ±7°. Past about 10° the flat DOM card reads as a flat sheet of paper being rotated rather than a card with substance, and the illusion inverts.

**Step 2: Isolate the blend context ahead of Task 4**

```scss
.tcg-card-frame {
  isolation: isolate; // contain the Task 4 `screen` blend to this card
}
```

Without `isolation: isolate`, `mix-blend-mode: screen` blends against whatever the nearest isolating ancestor turns out to be, and the sheen can leak onto the page background around the card. Set it now, while the cause-and-effect is easy to see.

**Step 3: Honour reduced motion for the animations that already exist**

`LandingView` currently runs `fadeUp` and `floatCard` unconditionally. Follow the precedent already set in `src/components/decks/DeckImportPanel.vue`:

```scss
@media (prefers-reduced-motion: reduce) {
  .hero-copy,
  article,
  .tcg-card-frame {
    animation: none;
  }
}
```

This is a pre-existing accessibility gap rather than something this feature introduces, but this feature is what makes it matter, so fix it here.

**Verify:** in a browser, the card is pixel-identical at rest and tilts smoothly toward the pointer with no jitter at the extremes. With reduced motion forced on, nothing on the page moves. Run `npm test` — the existing suite must not regress.

---

### Task 4: The foil component

**Files:**
- Create: `app/src/components/landing/HeroCardFoil.vue`

Props: `tilt: [number, number]`, `pointer: [number, number]`, `intensity?: number`. The component owns its canvas, renderer, and RAF loop, and nothing else in the app knows three.js exists.

**Step 1: Canvas placement**

```html
<canvas ref="canvas" class="hero-card-foil" aria-hidden="true" />
```

```scss
.hero-card-foil {
  position: absolute;
  inset: 0;
  z-index: 3;
  pointer-events: none;
  mix-blend-mode: screen;
  opacity: 0; // faded in once the first frame has rendered
  transition: opacity 600ms ease;
}
```

`z-index: 3` puts it above the header and footer (`z-index: 1`). That is intentional — foil sits over the print on a real card — but it makes the shader's text mask load-bearing rather than cosmetic. Start at `opacity: 0` and fade in after the first successful frame so a slow WebGL init never produces a visible pop.

**Step 2: Renderer setup**

```ts
const renderer = new THREE.WebGLRenderer({
  canvas: canvasEl,
  alpha: false,      // `screen` treats black as identity, so no alpha needed
  antialias: false,  // one fullscreen quad; there are no edges to alias
  // powerPreference is deliberately left at default: never wake the dGPU for a landing page
});
renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
```

Scene is one `PlaneGeometry(1, 1)` with a `ShaderMaterial` and an `OrthographicCamera`. Vertex shader bypasses the projection entirely:

```glsl
varying vec2 vUv;
void main() {
  vUv = uv;
  gl_Position = vec4(position.xy * 2.0, 0.0, 1.0);
}
```

`PlaneGeometry` positions span `-0.5..0.5`, so `* 2.0` fills clip space exactly.

**Step 3: The fragment shader**

Uniforms: `uTime`, `uTilt` (vec2), `uPointer` (vec2), `uIntensity` (float), `uArtboard` (vec4 UV rect), `uTextBand` (vec4 UV rect).

Target shape — tune the constants against the real card, they are a starting point rather than a result:

```glsl
precision highp float;

uniform float uTime;
uniform vec2  uTilt;      // -1..1, damped
uniform vec2  uPointer;   // 0..1, card-local
uniform float uIntensity;
uniform vec4  uArtboard;
uniform vec4  uTextBand;
varying vec2  vUv;

// Straight from src/styles/main.scss. No hue outside this ramp.
const vec3 EMBER = vec3(0.949, 0.627, 0.255); // #f2a041
const vec3 FLARE = vec3(0.941, 0.031, 0.463); // #f00876
const vec3 PARCH = vec3(1.000, 0.957, 0.929); // #fff4ed

float hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

float noise(vec2 p) {
  vec2 i = floor(p), f = fract(p);
  vec2 u = f * f * (3.0 - 2.0 * f);
  return mix(mix(hash(i),               hash(i + vec2(1.0, 0.0)), u.x),
             mix(hash(i + vec2(0.0,1.0)), hash(i + vec2(1.0, 1.0)), u.x), u.y);
}

float fbm(vec2 p) {
  return 0.5 * noise(p) + 0.25 * noise(p * 2.03) + 0.125 * noise(p * 4.01);
}

float inRect(vec2 uv, vec4 r, float feather) {
  vec2 a = smoothstep(r.xy - feather, r.xy + feather, uv);
  vec2 b = 1.0 - smoothstep(r.xy + r.zw - feather, r.xy + r.zw + feather, uv);
  return a.x * a.y * b.x * b.y;
}

void main() {
  vec2 uv = vUv;

  // Bevelled normal: edges lean outward so the rim catches light.
  vec2 edge = min(uv, 1.0 - uv);
  vec2 lean = (1.0 - smoothstep(0.0, 0.06, edge)) * sign(uv - 0.5);
  vec3 N = normalize(vec3(lean * 0.9, 1.0));
  vec3 V = normalize(vec3(uTilt * 0.65, 1.0));

  float fresnel = pow(1.0 - max(dot(N, V), 0.0), 3.0);

  // Diffraction: a diagonal band, noise-warped, swept by tilt plus slow drift.
  float diag  = dot(uv, normalize(vec2(1.0, 0.62)));
  float warp  = fbm(uv * 3.2 + uTime * 0.02);
  float sweep = diag * 2.6 + warp * 0.55 + dot(uTilt, vec2(0.8, -0.5)) + uTime * 0.015;
  float band  = 0.5 + 0.5 * sin(sweep * 6.2831);

  vec3 sheen = mix(EMBER, FLARE, smoothstep(0.25, 0.85, band));
  sheen = mix(sheen, PARCH, pow(band, 6.0) * 0.7);

  // Specular hotspot tracking the cursor.
  float spec = pow(1.0 - clamp(length((uv - uPointer) * vec2(1.0, 1.35)) * 1.6, 0.0, 1.0), 6.0);

  // Where the foil may sing, and where copy must stay readable.
  float zone = mix(0.30, 1.0, inRect(uv, uArtboard, 0.02))
             * mix(1.0, 0.28, inRect(uv, uTextBand, 0.03));

  float amount = (band * 0.16 + fresnel * 0.42 + spec * 0.55) * zone;

  // Dither. A warm dark palette bands visibly in 8-bit without it.
  amount += (hash(uv * 1024.0) - 0.5) * 0.006;

  gl_FragColor = vec4(sheen * amount * uIntensity, 1.0);
}
```

Two things here are not optional. The **constrained hue ramp** is what makes this read as vEDH rather than as a rainbow novelty — resist widening it, because full-spectrum foil is exactly the cliché this is trying to avoid. The **dither** looks like superstition until you see the banding on `#3a3431`, at which point it is obviously mandatory.

Author GLSL as plain template strings. No `glslify` and no Vite plugin — `vite.config.ts` stays untouched.

**Step 4: Mask measurement**

On mount and on every `ResizeObserver` fire against `.tcg-card-frame`, measure `.tcg-card-artboard` and the union of `.tcg-card-rulebox` + `.tcg-card-footer`, convert with `rectToUv`, and write the two vec4 uniforms. Also call `renderer.setSize` there. Measure in a `requestAnimationFrame` callback, not synchronously inside the observer, to avoid layout thrash.

**Step 5: The loop**

Single RAF. Each frame: advance `uTime` from a `THREE.Clock` delta, copy the smoothed tilt and pointer into uniforms, render. No per-frame allocation — reuse `Vector2` instances rather than assigning array literals to uniforms.

**Verify:** in a browser, the sheen sweeps as the pointer moves, stays strongest over the artboard, and leaves the rulebox and footer copy fully legible. Check the last point at reduced brightness and on a low-quality panel, not only on a good display.

---

### Task 5: Wire into LandingView behind a capability gate

**Files:**
- Modify: `app/src/views/LandingView.vue`

**Step 1: Async component boundary**

```ts
const HeroCardFoil = defineAsyncComponent(() => import('../components/landing/HeroCardFoil.vue'));
```

This is the boundary that keeps three.js out of the entry chunk. Task 1 Step 3 verifies it held.

**Step 2: The gate**

```ts
const foilEnabled = computed(() => supportsWebgl() && wideEnough.value && !saveData());
```

- `supportsWebgl()` — create a throwaway canvas, attempt `getContext('webgl2') ?? getContext('webgl')`, and discard it. Wrap in try/catch: some hardened browsers throw here rather than returning null.
- `wideEnough` — `>= 720px`, matching the existing media query so the fallback seam is already designed.
- `saveData()` — `navigator.connection?.saveData`. Respect it by never fetching the chunk.

Render with `v-if="foilEnabled"` inside `.tcg-card-frame`. Reduced motion is deliberately **not** in this gate: the foil still renders, just as a single static frame, because a static foil is prettier than no foil and costs nothing.

**Step 3: Bind tilt**

`usePointerTilt` is the single source of truth. `LandingView` binds it to the wrapper's CSS custom properties and passes the same values as props to `<HeroCardFoil>`. One value, two consumers — that is what keeps the sheen locked to the geometry.

**Verify:** with WebGL disabled in devtools, the page renders the unmodified CSS card and never requests the three chunk. Below 720px, likewise.

---

### Task 6: Lifecycle hardening

**Files:**
- Modify: `app/src/components/landing/HeroCardFoil.vue`

Each of these is a real failure mode observed with hero-canvas effects, not defensive padding.

**Step 1: Pause when invisible**

`IntersectionObserver` on the canvas: cancel the RAF on exit, resume on entry. Also listen for `visibilitychange` and stop while `document.hidden`. A landing page left open in a background tab must not burn GPU.

**Step 2: Reduced motion renders one frame**

When `prefers-reduced-motion: reduce` is set: set the frozen tilt, render exactly one frame, and never start the loop. Re-evaluate on the media query's `change` event so toggling the OS setting takes effect without a reload.

**Step 3: Handle context loss**

Listen for `webglcontextlost`: `preventDefault()`, stop the loop, and set the canvas to `opacity: 0` so the CSS card is left pristine. On `webglcontextrestored`, rebuild uniforms and resume. Context loss is routine on laptops that switch GPUs and on mobile Safari after backgrounding — without this, the card is left with a frozen sheen frame smeared across it.

**Step 4: Full teardown in `onUnmounted`**

Cancel RAF; disconnect both observers; remove listeners; `geometry.dispose()`, `material.dispose()`, `renderer.dispose()`, then `renderer.forceContextLoss()`. Navigating between `/` and `/login` repeatedly must not accumulate contexts — browsers reap them around ~16 and the symptom is other canvases on the site randomly going blank.

**Verify:** navigate `/` → `/login` → `/` twenty times and confirm via `chrome://gpu` or a heap snapshot that contexts and memory are flat. Background the tab and confirm the RAF stops.

---

### Task 7: Tests

**Files:**
- Create: `app/__tests__/HeroCardFoil.spec.ts`
- Modify: `app/e2e/create-and-join-game.spec.ts` or create `app/e2e/landing-foil.spec.ts`

**Step 1: The guarantee that actually matters**

jsdom has no WebGL, which makes it the perfect environment for asserting the fallback. Mount `LandingView` in jsdom and assert that the full CSS card renders — frame, artboard, all three rule lines, footer — and that no `canvas` is present. This is the regression test that protects the "everything is additive" principle, and it is worth more than any assertion about the shader.

**Step 2: Gate and lifecycle units**

`vi.mock('three')` to assert that the component constructs a renderer once, calls `setPixelRatio` with a value clamped to 2, and calls `dispose` plus `forceContextLoss` on unmount. Assert the gate returns false when `getContext` returns null and when it throws.

**Step 3: e2e smoke**

In Playwright, assert on `/` that the canvas exists with non-zero dimensions. Headless Chromium generally provides WebGL via SwiftShader; if it proves flaky in CI, assert the fallback path instead of adding retries — a flaky test here is worse than no test.

**Verify:** `npm test` and `npm run test:e2e` pass.

---

### Task 8: Performance and polish pass

**Files:** none — measurement only.

**Step 1: Budget check**

Record the `dist` delta. three plus this component should land around 150–170KB gzipped in the landing chunk. If it materially exceeds that, check for an accidental `three/addons` import — that is nearly always the cause.

**Step 2: Frame cost**

Chrome performance profile with the pointer moving continuously. The whole effect is one fullscreen quad and should sit in the low single-digit milliseconds per frame. If it does not, the shader is being run at an unclamped `devicePixelRatio` — re-check Task 4 Step 2.

**Step 3: Cross-browser QA**

- Safari: `mix-blend-mode` combined with transformed ancestors occasionally forces expensive layer compositing. Verify the tilt still holds 60fps on macOS Safari and iOS Safari specifically.
- Firefox: verify the blend composites against the card rather than the page background — this is the `isolation: isolate` from Task 3 Step 2 doing its job, and Firefox is where its absence shows first.

**Step 4: Taste pass**

Sit with it for a full minute without moving the pointer. If the idle drift is distracting, halve it. The failure mode for this kind of effect is being 30% too strong, and it is far easier to see that after a minute of not interacting than in the first five seconds of enthusiasm.

---

## Risks

| Risk | Mitigation |
| --- | --- |
| Sheen harms copy legibility | `uTextBand` mask damps to 0.28 over rulebox and footer; verified on a low-quality panel, not just a good one |
| Bundle regression on other routes | Async component boundary plus the Task 1 Step 3 chunk assertion |
| `mix-blend-mode` inconsistency across browsers | `isolation: isolate` on the frame; explicit Safari and Firefox QA |
| Battery drain | Default `powerPreference`, DPR clamped to 2, RAF paused when hidden or off-screen |
| Effect ends up too strong | Single `uIntensity` master uniform; deliberate cool-down taste pass in Task 8 |
| WebGL unavailable or context lost | Additive by construction — the CSS card is always the fallback, and context loss hides the canvas entirely |

## Rollback

Delete `HeroCardFoil.vue`, remove the `v-if` block from `LandingView.vue`, and drop the two dependencies. The Task 3 changes — tilt wrapper, `isolation`, reduced-motion handling — are worth keeping regardless, which is why they are sequenced as a standalone task.
