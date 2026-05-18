<template>
  <section class="landing">
    <div class="hero">
      <div class="hero-copy">
        <span class="eyebrow">vEDH multiplayer hub</span>
        <h1>Bring any TCG table online, instantly.</h1>
        <p>
          vEDH keeps matches clean and collaborative with live zones, shared counters, and
          flexible turn flow for any trading card game.
        </p>
        <div class="actions">
          <button class="primary" @click="goToSignup">Start a table</button>
          <button class="secondary" @click="goToLogin">Join a game</button>
        </div>
        <div class="hero-stats">
          <div>
            <strong>Live zones</strong>
            <span>Hands, boards, discard, and more.</span>
          </div>
          <div>
            <strong>Shared counters</strong>
            <span>Track life, energy, and custom tokens.</span>
          </div>
          <div>
            <strong>Turn flow</strong>
            <span>Pass priority, log moves, stay synced.</span>
          </div>
        </div>
      </div>
      <div class="hero-art" aria-hidden="true">
        <div class="hero-orbit orbit-a"></div>
        <div class="hero-orbit orbit-b"></div>
        <div class="hero-sigil sigil-top"></div>
        <div class="hero-sigil sigil-bottom"></div>

        <div class="tcg-card-shadow"></div>
        <div class="tcg-card-frame">
          <div class="tcg-card-edge"></div>
          <div class="tcg-card-header">
            <div>
              <span class="tcg-card-kicker">vector table relic</span>
              <strong>Mythic Junction // Table Nexus</strong>
            </div>
            <span class="tcg-card-cost">
              <span></span>
              <span></span>
              <span></span>
            </span>
          </div>

          <div class="tcg-card-artboard">
            <div class="art-grid"></div>
            <div class="art-ring ring-outer"></div>
            <div class="art-ring ring-inner"></div>
            <div class="art-node node-top"></div>
            <div class="art-node node-left"></div>
            <div class="art-node node-right"></div>
            <div class="art-node node-bottom"></div>
            <div class="art-glyph glyph-center"></div>
            <div class="art-glyph glyph-side"></div>
            <div class="art-ribbon ribbon-a"></div>
            <div class="art-ribbon ribbon-b"></div>
          </div>

          <div class="tcg-card-rulebox">
            <div class="rule-line">
              <span class="rule-icon diamond"></span>
              <span>Sync zones across every player board.</span>
            </div>
            <div class="rule-line">
              <span class="rule-icon pulse"></span>
              <span>Track counters, energy, shields, and momentum.</span>
            </div>
            <div class="rule-line">
              <span class="rule-icon link"></span>
              <span>Original vector hero card with crisp linework and no borrowed game framing.</span>
            </div>
          </div>

          <div class="tcg-card-footer">
            <div class="footer-badges">
              <span>zonecraft</span>
              <span>tempo</span>
              <span>lobby</span>
            </div>
            <div class="footer-stats">
              <span>07</span>
              <span>/</span>
              <span>11</span>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="feature-grid">
      <article>
        <h3>Table-state clarity</h3>
        <p>Track zones and permanents with a board that stays readable at a glance.</p>
      </article>
      <article>
        <h3>Match flow tools</h3>
        <p>Use phase markers, priority passes, and quick logs to keep play moving.</p>
      </article>
      <article>
        <h3>Deck workspace</h3>
        <p>Search, shuffle, reveal, and organize without leaving the table.</p>
      </article>
    </div>
    <div class="playbook">
      <div>
        <h2>Set up in seconds.</h2>
        <p>
          Create a lobby, invite your group, and play any ruleset you want. vEDH gives your
          table the structure you need without locking you into a single format.
        </p>
      </div>
      <div class="steps">
        <div>
          <span>01</span>
          <h4>Open a table</h4>
          <p>Name the match, pick player counts, and post the invite.</p>
        </div>
        <div>
          <span>02</span>
          <h4>Sync your decks</h4>
          <p>Shuffle, draw, and reveal cards with shared, live state.</p>
        </div>
        <div>
          <span>03</span>
          <h4>Play faster</h4>
          <p>Log moves and pass turns with friction-free controls.</p>
        </div>
      </div>
    </div>
    <div class="cta">
      <div>
        <h2>Ready to host your next match?</h2>
        <p>Keep every TCG table organized, even when the board gets wild.</p>
      </div>
      <button class="primary" @click="goToSignup">Create a table</button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';

const router = useRouter();

function goToSignup() {
  router.push({ name: 'signup' });
}

function goToLogin() {
  router.push({ name: 'login' });
}
</script>

<style scoped lang="scss">
.landing {
  --ink: var(--vedh-text);
  --muted: var(--vedh-muted);
  --accent: var(--vedh-primary);
  --accent-2: var(--vedh-secondary);
  --panel: rgba(74, 66, 63, 0.78);
  --panel-border: var(--vedh-border);
  display: grid;
  gap: clamp(2.5rem, 5vw, 5rem);
  padding: clamp(2.5rem, 6vw, 6.5rem) 0;
  position: relative;
  font-family: 'Space Grotesk', 'Segoe UI', sans-serif;
  color: var(--ink);
  isolation: isolate;
}

.landing::before,
.landing::after {
  content: '';
  position: fixed;
  inset: 0;
  width: 100vw;
  height: 100vh;
  z-index: -2;
  pointer-events: none;
}

.landing::before {
  background:
    radial-gradient(circle at 15% 20%, rgba(242, 160, 65, 0.24), transparent 60%),
    radial-gradient(circle at 85% 15%, rgba(240, 8, 118, 0.22), transparent 62%),
    radial-gradient(circle at 55% 80%, rgba(242, 160, 65, 0.12), transparent 65%),
    linear-gradient(160deg, rgba(58, 52, 49, 0.96), rgba(58, 52, 49, 0.68));
}

.landing::after {
  background:
    radial-gradient(circle at 40% -10%, rgba(33, 28, 26, 0.8), transparent 65%),
    radial-gradient(circle at 10% 110%, rgba(28, 24, 22, 0.75), transparent 70%),
    linear-gradient(180deg, rgba(58, 52, 49, 0.92), rgba(47, 41, 39, 0.78));
  z-index: -3;
}

.hero {
  display: grid;
  gap: clamp(2rem, 5vw, 4rem);
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  align-items: center;
}

.hero-copy {
  display: grid;
  gap: 1.25rem;
  animation: fadeUp 0.8s ease forwards;
}

.eyebrow {
  text-transform: uppercase;
  letter-spacing: 0.3em;
  font-size: 0.7rem;
  color: var(--muted);
}

.hero-stats {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
}

.hero-stats strong {
  display: block;
  font-size: 0.95rem;
}

.hero-stats span {
  color: var(--muted);
  font-size: 0.85rem;
}

.hero-art {
  position: relative;
  min-height: 490px;
  display: grid;
  place-items: center;
}

.hero-orbit,
.hero-sigil {
  position: absolute;
  pointer-events: none;
}

.hero-orbit {
  width: 420px;
  height: 420px;
  border-radius: 50%;
  border: 1px solid rgba(247, 247, 242, 0.08);
  opacity: 0.8;
}

.hero-orbit::before,
.hero-orbit::after {
  content: '';
  position: absolute;
  border-radius: 50%;
}

.orbit-a {
  transform: rotate(18deg);
  border-style: dashed;
  border-color: rgba(242, 160, 65, 0.24);
}

.orbit-a::before {
  inset: 34px;
  border: 1px solid rgba(242, 160, 65, 0.16);
}

.orbit-b {
  width: 320px;
  height: 320px;
  transform: rotate(-14deg);
  border-color: rgba(240, 8, 118, 0.22);
}

.orbit-b::after {
  inset: 22px;
  border: 1px dashed rgba(247, 247, 242, 0.12);
}

.hero-sigil {
  width: 68px;
  height: 68px;
  border: 1px solid rgba(247, 247, 242, 0.16);
  background: linear-gradient(145deg, rgba(242, 160, 65, 0.18), rgba(240, 8, 118, 0.12));
  clip-path: polygon(50% 0%, 100% 50%, 50% 100%, 0% 50%);
  box-shadow: 0 14px 32px rgba(0, 0, 0, 0.25);
}

.sigil-top {
  top: 18px;
  right: 12%;
}

.sigil-bottom {
  bottom: 24px;
  left: 10%;
  width: 54px;
  height: 54px;
}

.tcg-card-shadow {
  position: absolute;
  width: min(350px, 76vw);
  height: 470px;
  border-radius: 34px;
  background: radial-gradient(circle at 50% 50%, rgba(242, 160, 65, 0.22), rgba(58, 52, 49, 0));
  filter: blur(28px);
  transform: translateY(24px);
}

.tcg-card-frame {
  position: relative;
  width: min(360px, 76vw);
  min-height: 490px;
  padding: 1rem;
  border-radius: 30px;
  background:
    linear-gradient(155deg, rgba(255, 240, 225, 0.9), rgba(242, 160, 65, 0.28) 26%, rgba(240, 8, 118, 0.18) 54%, rgba(58, 52, 49, 0.94) 54%),
    linear-gradient(180deg, rgba(74, 66, 63, 0.94), rgba(47, 41, 39, 0.98));
  box-shadow: 0 32px 90px rgba(0, 0, 0, 0.42);
  overflow: hidden;
  animation: floatCard 7s ease-in-out infinite;
}

.tcg-card-edge {
  position: absolute;
  inset: 10px;
  border-radius: 24px;
  border: 1px solid rgba(247, 247, 242, 0.12);
  pointer-events: none;
}

.tcg-card-header,
.tcg-card-footer {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.tcg-card-header {
  padding: 0.25rem 0.3rem 0.9rem;
}

.tcg-card-header strong {
  display: block;
  font-size: 1.02rem;
  letter-spacing: 0.04em;
}

.tcg-card-kicker {
  display: block;
  text-transform: uppercase;
  letter-spacing: 0.22em;
  font-size: 0.58rem;
  color: rgba(247, 247, 242, 0.6);
  margin-bottom: 0.3rem;
}

.tcg-card-cost {
  display: inline-flex;
  gap: 0.35rem;
}

.tcg-card-cost span,
.footer-badges span {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.tcg-card-cost span {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 1px solid rgba(247, 247, 242, 0.28);
  background: linear-gradient(135deg, rgba(242, 160, 65, 0.5), rgba(240, 8, 118, 0.22));
}

.tcg-card-artboard {
  position: relative;
  height: 238px;
  border-radius: 22px;
  overflow: hidden;
  background:
    radial-gradient(circle at 50% 46%, rgba(242, 160, 65, 0.3), transparent 25%),
    linear-gradient(145deg, rgba(242, 160, 65, 0.22), rgba(47, 41, 39, 0.92) 58%, rgba(240, 8, 118, 0.22));
  border: 1px solid rgba(247, 247, 242, 0.12);
}

.art-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(247, 247, 242, 0.08) 1px, transparent 1px),
    linear-gradient(90deg, rgba(247, 247, 242, 0.08) 1px, transparent 1px);
  background-size: 28px 28px;
  mask-image: linear-gradient(180deg, rgba(0,0,0,0.8), transparent 100%);
}

.art-ring,
.art-node,
.art-glyph,
.art-ribbon {
  position: absolute;
}

.art-ring {
  border-radius: 50%;
  border: 1px solid rgba(247, 247, 242, 0.16);
}

.ring-outer {
  inset: 28px 42px;
}

.ring-inner {
  inset: 62px 82px;
  border-color: rgba(242, 160, 65, 0.28);
}

.art-node {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: linear-gradient(135deg, rgba(242, 160, 65, 0.95), rgba(240, 8, 118, 0.72));
  box-shadow: 0 0 0 6px rgba(242, 160, 65, 0.08);
}

.node-top { top: 30px; left: 50%; transform: translateX(-50%); }
.node-left { left: 36px; top: 50%; transform: translateY(-50%); }
.node-right { right: 36px; top: 50%; transform: translateY(-50%); }
.node-bottom { bottom: 28px; left: 50%; transform: translateX(-50%); }

.art-glyph {
  border: 1px solid rgba(247, 247, 242, 0.16);
  background: rgba(12, 14, 20, 0.3);
}

.glyph-center {
  width: 84px;
  height: 84px;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%) rotate(45deg);
}

.glyph-side {
  width: 42px;
  height: 42px;
  right: 54px;
  top: 38px;
  clip-path: polygon(50% 0%, 100% 100%, 0% 100%);
}

.art-ribbon {
  height: 2px;
  background: linear-gradient(90deg, transparent, rgba(247, 247, 242, 0.38), transparent);
}

.ribbon-a {
  left: 58px;
  right: 58px;
  top: 50%;
}

.ribbon-b {
  top: 42px;
  bottom: 42px;
  width: 2px;
  left: 50%;
  transform: translateX(-50%);
  background: linear-gradient(180deg, transparent, rgba(247, 247, 242, 0.38), transparent);
}

.tcg-card-rulebox {
  display: grid;
  gap: 0.65rem;
  padding: 1rem 0.25rem 0.9rem;
}

.rule-line {
  display: grid;
  grid-template-columns: 18px 1fr;
  gap: 0.7rem;
  align-items: center;
  color: rgba(247, 247, 242, 0.84);
  font-size: 0.88rem;
  line-height: 1.45;
}

.rule-icon {
  width: 18px;
  height: 18px;
  display: inline-block;
  position: relative;
}

.rule-icon.diamond {
  clip-path: polygon(50% 0%, 100% 50%, 50% 100%, 0% 50%);
  background: linear-gradient(135deg, rgba(242, 160, 65, 0.95), rgba(240, 8, 118, 0.72));
}

.rule-icon.pulse::before,
.rule-icon.link::before,
.rule-icon.link::after {
  content: '';
  position: absolute;
}

.rule-icon.pulse::before {
  inset: 3px;
  border-radius: 50%;
  border: 2px solid rgba(240, 8, 118, 0.86);
  box-shadow: 0 0 0 3px rgba(240, 8, 118, 0.14);
}

.rule-icon.link::before,
.rule-icon.link::after {
  width: 8px;
  height: 12px;
  border: 2px solid rgba(242, 160, 65, 0.82);
  border-radius: 999px;
  top: 2px;
}

.rule-icon.link::before { left: 0; }
.rule-icon.link::after { right: 0; }

.tcg-card-footer {
  padding: 0.65rem 0.2rem 0.15rem;
  border-top: 1px solid rgba(247, 247, 242, 0.08);
}

.footer-badges {
  display: flex;
  gap: 0.45rem;
  flex-wrap: wrap;
}

.footer-badges span {
  padding: 0.26rem 0.55rem;
  border-radius: 999px;
  border: 1px solid rgba(247, 247, 242, 0.12);
  background: rgba(247, 247, 242, 0.05);
  text-transform: uppercase;
  letter-spacing: 0.12em;
  font-size: 0.55rem;
}

.footer-stats {
  display: inline-flex;
  gap: 0.25rem;
  color: rgba(247, 247, 242, 0.68);
  font-family: 'Space Mono', monospace;
  font-size: 0.85rem;
}

h1 {
  font-size: clamp(2.7rem, 5.5vw, 4rem);
  margin: 0;
}

p {
  margin: 0;
  line-height: 1.6;
  color: var(--muted);
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

button {
  cursor: pointer;
  border: none;
  border-radius: 999px;
  padding: 0.75rem 1.8rem;
  font-weight: 600;
  font-size: 1rem;
}

button.primary {
  background: linear-gradient(120deg, var(--accent), var(--accent-2));
  color: #111;
}

button.secondary {
  background: rgba(255, 255, 255, 0.1);
  color: var(--ink);
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 1.5rem;
}

article {
  background: var(--panel);
  border: 1px solid var(--panel-border);
  border-radius: 16px;
  padding: 1.5rem;
  backdrop-filter: blur(12px);
  animation: fadeUp 0.8s ease forwards;
}

article:nth-child(2) {
  animation-delay: 0.1s;
}

article:nth-child(3) {
  animation-delay: 0.2s;
}

h3 {
  margin-top: 0;
}

.playbook {
  display: grid;
  gap: 2rem;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  align-items: start;
  background: rgba(10, 10, 14, 0.65);
  border: 1px solid var(--panel-border);
  border-radius: 24px;
  padding: clamp(1.5rem, 4vw, 2.5rem);
}

.steps {
  display: grid;
  gap: 1.25rem;
}

.steps span {
  font-size: 0.75rem;
  letter-spacing: 0.2em;
  color: var(--accent);
  text-transform: uppercase;
}

.steps h4 {
  margin: 0.35rem 0 0.2rem;
}

.steps p {
  font-size: 0.95rem;
}

.cta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
  padding: clamp(1.5rem, 4vw, 2.75rem);
  border-radius: 28px;
  background: linear-gradient(140deg, rgba(242, 160, 65, 0.22), rgba(240, 8, 118, 0.14));
  border: 1px solid rgba(255, 255, 255, 0.12);
}

.cta h2 {
  margin: 0 0 0.5rem;
}

@keyframes fadeUp {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes floatCard {
  0%,
  100% {
    transform: translateY(0) rotate(-1.5deg);
  }
  50% {
    transform: translateY(-10px) rotate(1deg);
  }
}

@media (max-width: 720px) {
  .hero-art {
    min-height: 320px;
  }

  .hero-orbit {
    width: 300px;
    height: 300px;
  }

  .orbit-b {
    width: 240px;
    height: 240px;
  }

  .tcg-card-shadow {
    width: min(280px, 78vw);
    height: 380px;
  }

  .tcg-card-frame {
    width: min(280px, 78vw);
    min-height: 390px;
    padding: 0.8rem;
  }

  .tcg-card-artboard {
    height: 180px;
  }

  .rule-line {
    font-size: 0.8rem;
  }

  .cta {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
