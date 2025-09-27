<template>
  <section class="games">
    <header class="page-header">
      <div>
        <h1>Your games</h1>
        <p>Browse existing games or spin up a new table.</p>
      </div>
      <button class="primary" @click="showCreateModal = true">Create game</button>
    </header>

    <div v-if="loading" class="skeleton-grid">
      <div v-for="n in 4" :key="n" class="skeleton-card" />
    </div>

  <div v-else class="games-grid">
  <article v-for="game in games" :key="game.ID" @click="openGame(game.ID)">
        <h3>{{ formatGameTitle(game) }}</h3>
        <ul>
          <li v-for="player in game.Players" :key="player.ID || player.Username">{{ player.Username }}</li>
        </ul>
        <footer>
          <span>{{ formatTurn(game.Turn) }}</span>
        </footer>
      </article>
    </div>

    <FormCreateGame v-if="showCreateModal" @close="showCreateModal = false" @created="handleGameCreated" />
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useGamesStore } from '../stores/games';
import FormCreateGame from '../components/games/FormCreateGame.vue';

const gamesStore = useGamesStore();
const router = useRouter();

const loading = computed(() => gamesStore.loading);
const games = computed(() => gamesStore.games);
const showCreateModal = ref(false);

onMounted(() => {
  gamesStore.fetchGames();
});

function openGame(id: string) {
  router.push({ name: 'board', params: { id } });
}

function handleGameCreated(id: string) {
  showCreateModal.value = false;
  router.push({ name: 'board', params: { id } });
}

function formatGameTitle(game: { ID: string; Players: { Username: string }[] }) {
  const playerNames = game.Players.map((player) => player.Username).join(', ');
  return playerNames || `Game ${game.ID.slice(0, 4)}`;
}

function formatTurn(turn?: { Player?: string; Phase?: string; Number?: number }) {
  if (!turn) return '—';
  return `${turn.Player ?? 'Unknown'} • ${turn.Phase ?? 'Phase'} • Turn ${turn.Number ?? '-'}`;
}
</script>

<style scoped lang="scss">
.games {
  display: grid;
  gap: clamp(1.5rem, 3vw, 2.5rem);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.games-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1.25rem;
}

article {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  padding: 1.5rem;
  cursor: pointer;
  transition: transform 0.15s ease, border-color 0.15s ease;
}

article:hover {
  transform: translateY(-4px);
  border-color: rgba(133, 215, 255, 0.6);
}

button.primary {
  background: linear-gradient(120deg, #85d7ff, #26c2ff);
  border: none;
  border-radius: 999px;
  padding: 0.65rem 1.4rem;
  font-weight: 600;
  cursor: pointer;
}

.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1.25rem;
}

.skeleton-card {
  height: 140px;
  border-radius: 16px;
  background: linear-gradient(110deg, rgba(255,255,255,0.06) 25%, rgba(255,255,255,0.12) 37%, rgba(255,255,255,0.06) 63%);
  background-size: 400% 100%;
  animation: shimmer 1.6s ease infinite;
}

@keyframes shimmer {
  0% { background-position: 0% 0%; }
  100% { background-position: -135% 0%; }
}
</style>
