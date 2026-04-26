<template>
  <section class="empty-state">
    <h1>Game not found</h1>
    <p>The game you were looking for doesn’t exist or may have been closed.</p>
    <RouterLink class="primary" to="/games">Return to games</RouterLink>
  </section>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router';
</script>

<style scoped lang="scss">
.empty-state {
  min-height: 60vh;
  display: grid;
  place-items: center;
  gap: 1rem;
  text-align: center;
}

.primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  padding: 0.75rem 1.5rem;
  background: linear-gradient(120deg, #85d7ff, #26c2ff);
  color: #0b1016;
  text-decoration: none;
  font-weight: 600;
}
</style>
