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
        <div class="card-stack back"></div>
        <div class="card-stack mid"></div>
        <div class="card-stack front">
          <div class="chip">+2</div>
          <div class="chip">Shield</div>
          <div class="chip">Energy</div>
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
  --ink: #f7f7f2;
  --muted: rgba(247, 247, 242, 0.7);
  --accent: #f7b500;
  --accent-2: #ff6b6b;
  --panel: rgba(18, 18, 22, 0.75);
  --panel-border: rgba(247, 247, 242, 0.08);
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
    radial-gradient(circle at 15% 20%, rgba(247, 181, 0, 0.22), transparent 60%),
    radial-gradient(circle at 85% 15%, rgba(255, 107, 107, 0.2), transparent 62%),
    radial-gradient(circle at 55% 80%, rgba(80, 190, 255, 0.16), transparent 65%),
    linear-gradient(160deg, rgba(10, 12, 18, 0.95), rgba(10, 12, 18, 0.6));
}

.landing::after {
  background:
    radial-gradient(circle at 40% -10%, rgba(8, 8, 12, 0.85), transparent 65%),
    radial-gradient(circle at 10% 110%, rgba(12, 12, 18, 0.8), transparent 70%),
    linear-gradient(180deg, rgba(10, 12, 18, 0.92), rgba(10, 12, 18, 0.75));
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
  min-height: 320px;
  display: grid;
  place-items: center;
}

.card-stack {
  position: absolute;
  width: min(320px, 70vw);
  height: 420px;
  border-radius: 26px;
  background: linear-gradient(150deg, rgba(248, 232, 195, 0.12), rgba(255, 255, 255, 0));
  border: 1px solid rgba(255, 255, 255, 0.14);
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.35);
  backdrop-filter: blur(12px);
  animation: float 6s ease-in-out infinite;
}

.card-stack.back {
  transform: translate(-28px, -30px) rotate(-6deg);
  background: linear-gradient(140deg, rgba(80, 190, 255, 0.18), rgba(20, 20, 30, 0.2));
}

.card-stack.mid {
  transform: translate(8px, -10px) rotate(4deg);
  animation-delay: 0.4s;
}

.card-stack.front {
  position: relative;
  display: grid;
  align-content: center;
  gap: 0.6rem;
  padding: 1.5rem;
  background: linear-gradient(160deg, rgba(247, 181, 0, 0.2), rgba(20, 20, 30, 0.25));
  animation-delay: 0.8s;
}

.chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.4rem 0.8rem;
  border-radius: 999px;
  background: rgba(12, 12, 16, 0.78);
  border: 1px solid rgba(255, 255, 255, 0.16);
  font-weight: 600;
  width: fit-content;
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
  background: linear-gradient(140deg, rgba(247, 181, 0, 0.2), rgba(255, 107, 107, 0.12));
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

@keyframes float {
  0%,
  100% {
    transform: translate(8px, -10px) rotate(4deg);
  }
  50% {
    transform: translate(12px, -18px) rotate(6deg);
  }
}

@media (max-width: 720px) {
  .hero-art {
    min-height: 260px;
  }

  .card-stack {
    height: 340px;
  }

  .cta {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
