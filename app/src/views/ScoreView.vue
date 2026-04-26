<template>
  <section class="score">
    <header>
      <h1>Scoreboard</h1>
      <p>Track commander damage and life totals across the table.</p>
    </header>
    <div class="score-grid">
      <article v-for="player in game?.Players ?? []" :key="player.ID">
        <h2>{{ player.Username }}</h2>
        <p class="life">{{ player.Boardstate?.Life ?? '—' }} life</p>
        <div class="commander-damage">
          <h3>Commander damage</h3>
          <p class="muted">Coming soon in v2</p>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useGamesStore } from '../stores/games';

const games = useGamesStore();
const game = computed(() => games.activeGame);
</script>

<style scoped lang="scss">
.score {
  display: grid;
  gap: 1.5rem;
}

.score-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1.5rem;
}

article {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 1.5rem;
}

.life {
  font-size: 1.5rem;
  margin: 0.5rem 0;
}

.muted {
  color: rgba(255, 255, 255, 0.6);
}
</style>
