<template>
  <section class="games">
    <header class="page-header">
      <div>
        <h1>Your games</h1>
        <p>Browse existing games or spin up a new table.</p>
      </div>
      <div class="header-actions">
        <div class="search">
          <input
            v-model="searchQuery"
            type="search"
            inputmode="search"
            placeholder="Search games…"
            aria-label="Search games by name or ID"
          />
          <button v-if="searchQuery" class="clear" @click="searchQuery = ''" aria-label="Clear search">×</button>
        </div>
        <button class="primary" @click="showCreateModal = true">Create game</button>
      </div>
    </header>

    <!-- Loading skeleton -->
    <div v-if="loading" class="skeleton-table">
      <div class="skeleton-row" v-for="n in 4" :key="n">
        <div class="sk sk-title" />
        <div class="sk sk-players" />
        <div class="sk sk-turn" />
        <div class="sk sk-created" />
        <div class="sk sk-actions" />
      </div>
    </div>

    <!-- Detail list -->
    <div v-else class="table-container">
      <div v-if="!filteredGames.length" class="empty" role="status">
        <p v-if="searchQuery">No games match “{{ searchQuery }}”.</p>
        <p v-else>No games to show yet.</p>
      </div>
      <table class="games-table">
        <thead>
          <tr>
            <th scope="col">Game</th>
            <th scope="col">Players</th>
            <th scope="col">Turn</th>
            <th scope="col">Created</th>
            <th scope="col" class="actions-col">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="game in filteredGames" :key="game.ID" @click="openGame(game.ID)" class="clickable">
            <td data-label="Game">
              <div class="cell-title">{{ formatGameTitle(game) }}</div>
              <div class="cell-sub">ID: {{ game.ID }}</div>
            </td>
            <td data-label="Players">
              <div class="cell-players">
                <span v-if="game.Players?.length === 0">—</span>
                <template v-else>
                  <span v-for="(player, i) in game.Players" :key="player.ID || player.Username">
                    {{ player.Username }}<span v-if="i < game.Players.length - 1">, </span>
                  </span>
                </template>
              </div>
            </td>
            <td data-label="Turn">
              <div class="cell-turn">{{ formatTurn(game.Turn) }}</div>
            </td>
            <td data-label="Created">
              <div class="cell-created">{{ formatDate(game.CreatedAt) }}</div>
            </td>
            <td class="actions" @click.stop>
              <button class="link" @click="openGame(game.ID)">Open</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <FormCreateGame v-if="showCreateModal" @close="showCreateModal = false" @created="handleGameCreated" />
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useGamesStore } from '../stores/games';
import { useAuthStore } from '../stores/auth';
import FormCreateGame from '../components/games/FormCreateGame.vue';

const gamesStore = useGamesStore();
const auth = useAuthStore();
const router = useRouter();

const loading = computed(() => gamesStore.loading);
const games = computed(() => gamesStore.games);
const showCreateModal = ref(false);
const searchQuery = ref('');

const filteredGames = computed(() => {
  const list = games.value ?? [];
  const profile = auth.profile;
  const username = profile?.Username?.toLowerCase();
  const userId = profile?.ID;

  // Only show games the current authenticated user is in.
  const userGames = (!userId && !username)
    ? list
    : list.filter((g) => (g.Players ?? []).some((p) => {
      if (userId && p?.ID && p.ID === userId) return true;
      if (username && p?.Username) return p.Username.toLowerCase() === username;
      return false;
    }));

  const q = searchQuery.value.trim().toLowerCase();
  if (!q) return userGames;
  return userGames.filter((g) => {
    const title = formatGameTitle(g).toLowerCase();
    return title.includes(q) || g.ID.toLowerCase().includes(q);
  });
});

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

function formatGameTitle(game: { ID: string; Players?: { Username?: string }[] }) {
  const playerNames = (game.Players ?? []).map((player) => player?.Username).filter(Boolean).join(', ');
  return playerNames || `Game ${game.ID.slice(0, 4)}`;
}

function formatTurn(turn?: { Player?: string; Phase?: string; Number?: number }) {
  if (!turn) return '—';
  return `${turn.Player ?? 'Unknown'} • ${turn.Phase ?? 'Phase'} • Turn ${turn.Number ?? '-'}`;
}

function formatDate(iso?: string) {
  if (!iso) return '—';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '—';
  return d.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
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

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.search {
  position: relative;
}

.search input[type='search'] {
  appearance: none;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: #fff;
  padding: 0.55rem 2rem 0.55rem 0.9rem;
  border-radius: 10px;
  outline: none;
}

.search .clear {
  position: absolute;
  right: 0.25rem;
  top: 50%;
  transform: translateY(-50%);
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.8);
  cursor: pointer;
}

.table-container {
  width: 100%;
  overflow-x: auto;
}

.empty {
  padding: 1rem;
  opacity: 0.8;
}

.games-table {
  width: 100%;
  border-collapse: collapse;
  border-spacing: 0;
}

.games-table thead th {
  text-align: left;
  font-weight: 600;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.12);
  color: rgba(255, 255, 255, 0.85);
}

.games-table tbody td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  vertical-align: middle;
}

.games-table tbody tr.clickable {
  cursor: pointer;
  transition: background 0.15s ease;
}

.games-table tbody tr.clickable:hover {
  background: rgba(255, 255, 255, 0.04);
}

.cell-title {
  font-weight: 600;
}

.cell-sub {
  opacity: 0.7;
  font-size: 0.85rem;
}

.actions-col,
.actions {
  white-space: nowrap;
  width: 1%;
}

button.primary {
  background: linear-gradient(120deg, #85d7ff, #26c2ff);
  border: none;
  border-radius: 999px;
  padding: 0.65rem 1.4rem;
  font-weight: 600;
  cursor: pointer;
}

.skeleton-table {
  display: grid;
  gap: 0.5rem;
}

.skeleton-row {
  display: grid;
  grid-template-columns: 2fr 2fr 1.5fr 1fr auto;
  gap: 0.75rem;
  align-items: center;
}

.sk {
  height: 18px;
  border-radius: 6px;
  background: linear-gradient(110deg, rgba(255,255,255,0.06) 25%, rgba(255,255,255,0.12) 37%, rgba(255,255,255,0.06) 63%);
  background-size: 400% 100%;
  animation: shimmer 1.6s ease infinite;
}

.sk-title { width: 60%; }
.sk-players { width: 80%; }
.sk-turn { width: 70%; }
.sk-created { width: 50%; }
.sk-actions { width: 64px; }

@keyframes shimmer {
  0% { background-position: 0% 0%; }
  100% { background-position: -135% 0%; }
}

/* Responsive: stack cells on narrow screens */
@media (max-width: 720px) {
  .games-table thead { display: none; }
  .games-table, .games-table tbody, .games-table tr, .games-table td { display: block; width: 100%; }
  .games-table tbody tr { border: 1px solid rgba(255,255,255,0.08); border-radius: 12px; margin-bottom: 0.75rem; overflow: hidden; }
  .games-table tbody td { display: grid; grid-template-columns: 9ch 1fr; gap: 0.5rem; }
  .games-table tbody td::before {
    content: attr(data-label);
    font-weight: 600;
    color: rgba(255,255,255,0.8);
  }
  .actions { display: flex; justify-content: flex-end; }
}

button.link {
  background: transparent;
  border: none;
  color: #85d7ff;
  font-weight: 600;
  cursor: pointer;
}
</style>
