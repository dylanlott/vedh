<template>
  <section class="score">
    <header>
      <h1>Scoreboard</h1>
      <p>Track commander damage and life totals across the table.</p>
    </header>
    <div class="score-grid">
      <article v-for="player in game?.Players ?? []" :key="player.ID">
        <div class="player-head">
          <h2>{{ player.Username }}</h2>
          <span class="life">{{ player.Boardstate?.Life ?? '—' }} life</span>
        </div>
        <div class="score-meta">
          <span class="chip">Battlefield {{ player.Boardstate?.Battlefield?.length ?? 0 }}</span>
          <span class="chip">Hand {{ player.Boardstate?.Hand?.length ?? 0 }}</span>
          <span class="chip">GY {{ player.Boardstate?.Graveyard?.length ?? 0 }}</span>
        </div>
        <div class="commander-damage">
          <h3>Commander damage</h3>
          <p class="muted">Coming soon in v2</p>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import { useGamesStore } from '../stores/games';

const games = useGamesStore();
const auth = useAuthStore();
const route = useRoute();
const game = computed(() => games.activeGame);

onMounted(async () => {
  const gameID = route.params.id as string;
  if (gameID) {
    await games.loadGame(gameID, auth.profile?.ID);
  }
});

watch([
  () => auth.profile?.ID,
  () => route.params.id,
], ([userID, gameID]) => {
  if (typeof gameID === 'string' && gameID && typeof userID === 'string' && userID) {
    games.subscribeToGame(gameID, userID);
  }
});

onBeforeUnmount(() => {
  games.clearActiveGame();
});
</script>

<style scoped lang="scss">
.score {
  display: grid;
  gap: 1.5rem;
}

.score > header {
  display: grid;
  gap: 0.35rem;
}

.score > header p {
  margin: 0;
  color: var(--vedh-muted);
}

.score-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1.5rem;
}

article {
  background: var(--vedh-panel);
  border: 1px solid var(--vedh-border);
  border-radius: 18px;
  padding: 1.5rem;
  box-shadow: 0 16px 34px rgba(21, 12, 9, 0.18);
  display: grid;
  gap: 0.9rem;
}

.player-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.player-head h2,
.commander-damage h3 {
  margin: 0;
}

.life {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.35rem 0.7rem;
  background: rgba(var(--vedh-primary-rgb), 0.14);
  border: 1px solid rgba(var(--vedh-primary-rgb), 0.24);
  color: var(--vedh-text);
  font-size: 1rem;
  font-weight: 600;
}

.score-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
}

.chip {
  font-size: 0.74rem;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  padding: 0.24rem 0.55rem;
  border-radius: 999px;
  background: rgba(255, 244, 237, 0.07);
  border: 1px solid var(--vedh-border);
  color: var(--vedh-muted);
}

.muted {
  color: var(--vedh-muted);
  margin: 0.25rem 0 0;
}
</style>
