<template>
  <section class="landing">
    <div class="hero">
      <h1>Commander games, streamlined.</h1>
      <p>
        Track life totals, board states, and the stack in real time. EDH Go v2 is rebuilt with Vue 3 for the modern commander table.
      </p>
      <div class="actions">
        <button class="primary" @click="goToSignup">Get started</button>
        <button class="secondary" @click="goToLogin">Log in</button>
      </div>
    </div>
    <div class="feature-grid">
      <article>
        <h3>Realtime updates</h3>
        <p>GraphQL subscriptions keep every player in sync as turns unfold.</p>
      </article>
      <article>
        <h3>Board control</h3>
        <p>Move cards between zones with a drag-and-drop interface tuned for commander.</p>
      </article>
      <article>
        <h3>Deck insights</h3>
        <p>Search the card database and build your library right in the browser.</p>
      </article>
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
  display: grid;
  gap: clamp(2rem, 4vw, 4rem);
  padding: clamp(2rem, 6vw, 6rem) 0;
}

.hero {
  display: grid;
  gap: 1.25rem;
  max-width: 680px;
}

h1 {
  font-size: clamp(2.5rem, 5vw, 3.5rem);
  margin: 0;
}

p {
  margin: 0;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.78);
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
  padding: 0.75rem 1.75rem;
  font-weight: 600;
  font-size: 1rem;
}

button.primary {
  background: linear-gradient(120deg, #85d7ff, #26c2ff);
  color: #0b1016;
}

button.secondary {
  background: rgba(255, 255, 255, 0.12);
  color: #f5f5f5;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 1.5rem;
}

article {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  padding: 1.5rem;
  backdrop-filter: blur(12px);
}

h3 {
  margin-top: 0;
}
</style>
