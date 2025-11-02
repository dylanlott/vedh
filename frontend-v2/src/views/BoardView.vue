<template>
  <section class="board" v-if="game">
    <header class="board-header">
      <h1>Game {{ game.ID }}</h1>
      <p>Turn {{ game.Turn?.Number ?? '—' }} • {{ game.Turn?.Player ?? 'Unknown player' }} ({{ game.Turn?.Phase ?? 'Phase' }})</p>
    </header>

    <div class="board-grid">
      <!-- Stack first (compact) -->
      <section class="stack">
        <header>
          <h2>Stack</h2>
        </header>
        <ol>
          <li v-for="card in game.Stack" :key="card.ID" class="card">{{ card.Name }}</li>
        </ol>
      </section>

      <!-- Players vertical list -->
      <aside class="players">
        <article v-for="player in game.Players" :key="player.ID" :class="{ active: isActivePlayer(player.Username) }">
          <header>
            <h2>{{ player.Username }}</h2>
            <span class="life">{{ player.Boardstate?.Life ?? '—' }} life</span>
            <!-- Toolbar only for self -->
            <nav v-if="isSelf(player.Username)" class="player-toolbar">
              <button
                class="tool"
                title="Draw 1"
                :disabled="(player.Boardstate?.Library?.length ?? 0) === 0"
                @click="draw(player.Username)"
              >🎴 Draw</button>
              <button
                class="tool"
                title="Mill 1"
                :disabled="(player.Boardstate?.Library?.length ?? 0) === 0"
                @click="mill(player.Username)"
              >🗑️ Mill</button>
              <button
                class="tool"
                title="Reveal top of library"
                :disabled="(player.Boardstate?.Library?.length ?? 0) === 0"
                @click="revealTop(player.Username)"
              >👁️ Reveal top</button>
              <button
                class="tool"
                title="Scry 1"
                :disabled="(player.Boardstate?.Library?.length ?? 0) === 0"
                @click="scryOne(player.Username)"
              >🔮 Scry 1</button>
              <button
                class="tool"
                title="Shuffle library"
                :disabled="(player.Boardstate?.Library?.length ?? 0) < 2"
                @click="shuffleLibrary(player.Username)"
              >🔀 Shuffle</button>
              <div class="life-tools">
                <button class="tool" title="Lose 1 life" @click="changeLife(player.Username, -1)">−1</button>
                <button class="tool" title="Gain 1 life" @click="changeLife(player.Username, 1)">+1</button>
              </div>
            </nav>
          </header>
          <div class="zone" :data-zone="'Commander'" @dragover.prevent @drop.prevent="onDrop(player.Username, 'Commander')">
            <h3>Commander</h3>
            <ul>
                <li
                  v-for="card in player.Boardstate?.Commander ?? []"
                  :key="card.ID"
                  class="card"
                  draggable="true"
                  @dragstart="onDragStart(card, player.Username, 'Commander')"
                  @click="quickMove(card, player.Username, 'Commander')"
                >
                  {{ card.Name }}
                </li>
            </ul>
          </div>
          <div class="zone" :data-zone="'Battlefield'" @dragover.prevent @drop.prevent="onDrop(player.Username, 'Battlefield')">
            <h3>Battlefield</h3>
            <ul>
                <li v-for="card in player.Boardstate?.Battlefield ?? []" :key="card.ID" class="card" draggable="true" @dragstart="onDragStart(card, player.Username, 'Battlefield')" @click="quickMove(card, player.Username, 'Battlefield')">
                  {{ card.Name }}
                </li>
            </ul>
          </div>
          <div class="zone" :data-zone="'Hand'" @dragover.prevent @drop.prevent="onDrop(player.Username, 'Hand')">
            <h3>Hand ({{ player.Boardstate?.Hand?.length ?? 0 }})</h3>
            <ul>
              <li
                v-for="card in player.Boardstate?.Hand ?? []"
                :key="card.ID"
                class="card"
                draggable="true"
                @dragstart="onDragStart(card, player.Username, 'Hand')"
                @click="quickMove(card, player.Username, 'Hand')"
              >
                {{ card.Name }}
              </li>
            </ul>
          </div>
          <div class="zone" :data-zone="'Graveyard'" @dragover.prevent @drop.prevent="onDrop(player.Username, 'Graveyard')">
            <h3>Graveyard ({{ player.Boardstate?.Graveyard?.length ?? 0 }})</h3>
            <ul>
                <li
                  v-for="card in player.Boardstate?.Graveyard ?? []"
                  :key="card.ID"
                  class="card"
                  draggable="true"
                  @dragstart="onDragStart(card, player.Username, 'Graveyard')"
                  @click="quickMove(card, player.Username, 'Graveyard')"
                >
                  {{ card.Name }}
                </li>
            </ul>
          </div>
          <div class="zone" :data-zone="'Exiled'" @dragover.prevent @drop.prevent="onDrop(player.Username, 'Exiled')">
            <h3>Exiled ({{ player.Boardstate?.Exiled?.length ?? 0 }})</h3>
            <ul>
                <li
                  v-for="card in player.Boardstate?.Exiled ?? []"
                  :key="card.ID"
                  class="card"
                  draggable="true"
                  @dragstart="onDragStart(card, player.Username, 'Exiled')"
                  @click="quickMove(card, player.Username, 'Exiled')"
                >
                  {{ card.Name }}
                </li>
            </ul>
          </div>
          <div class="zone" :data-zone="'Revealed'" @dragover.prevent @drop.prevent="onDrop(player.Username, 'Revealed')">
            <h3>Revealed ({{ player.Boardstate?.Revealed?.length ?? 0 }})</h3>
            <ul>
                <li
                  v-for="card in player.Boardstate?.Revealed ?? []"
                  :key="card.ID"
                  class="card"
                  draggable="true"
                  @dragstart="onDragStart(card, player.Username, 'Revealed')"
                  @click="quickMove(card, player.Username, 'Revealed')"
                >
                  {{ card.Name }}
                </li>
            </ul>
          </div>
          <div class="zone" :data-zone="'Controlled'" @dragover.prevent @drop.prevent="onDrop(player.Username, 'Controlled')">
            <h3>Controlled ({{ player.Boardstate?.Controlled?.length ?? 0 }})</h3>
            <ul>
                <li
                  v-for="card in player.Boardstate?.Controlled ?? []"
                  :key="card.ID"
                  class="card"
                  draggable="true"
                  @dragstart="onDragStart(card, player.Username, 'Controlled')"
                  @click="quickMove(card, player.Username, 'Controlled')"
                >
                  {{ card.Name }}
                </li>
            </ul>
          </div>
          <div class="zone" :data-zone="'Library'" @dragover.prevent @drop.prevent="onDrop(player.Username, 'Library')">
            <h3>
              Library ({{ player.Boardstate?.Library?.length ?? 0 }})
              <small v-if="isSelf(player.Username) && (player.Boardstate?.Library?.length ?? 0) > 0" class="muted">
                • Top: {{ player.Boardstate?.Library?.[0]?.Name ?? '—' }}
              </small>
            </h3>
            <ul class="library">
              <li v-if="isSelf(player.Username) && (player.Boardstate?.Library?.length ?? 0) === 0" class="card muted">Empty</li>
              <li v-else-if="!isSelf(player.Username)" class="card muted">Hidden</li>
            </ul>
          </div>
        </article>
      </aside>

      
    </div>
  </section>
  <section v-else class="loading-state">
    <p>Loading game…</p>
  </section>

  <!-- Scry 1 modal (self-only) -->
  <div v-if="scry?.open && isSelf(scry?.username)" class="scry-overlay">
    <div class="scry-modal">
      <header>Scry 1</header>
      <p v-if="scry?.card">Top card: <strong>{{ scry.card.Name }}</strong></p>
      <div class="scry-actions">
        <button class="tool" @click="scryKeepTop">Keep on top</button>
        <button class="tool" @click="scryPutBottom">Put on bottom</button>
      </div>
    </div>
  </div>

  <!-- Toasts -->
  <div class="toasts">
    <div class="toast" v-for="t in toasts" :key="t.id">{{ t.text }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useGamesStore } from '../stores/games';
import { useAuthStore } from '../stores/auth';
import { apolloClient } from '../services/apollo';
import { UPDATE_BOARDSTATE_MUTATION } from '../graphql/mutations';

const games = useGamesStore();
const auth = useAuthStore();
const route = useRoute();

// Zone typing shared across helpers
const zones = ['Commander','Battlefield','Hand','Graveyard','Exiled','Revealed','Library','Controlled'] as const;
type Zone = typeof zones[number];

const game = computed(() => games.activeGame);

onMounted(async () => {
  const gameID = route.params.id as string;
  await games.loadGame(gameID, auth.profile?.ID);
});

onBeforeUnmount(() => {
  games.clearActiveGame();
});

function isActivePlayer(username: string) {
  return username === game.value?.Turn?.Player;
}

function isSelf(username: string) {
  return username === auth.profile?.Username;
}

// Basic drag-and-drop state
const dragged = ref<{ card: { ID: string; Name: string }; fromUser: string; fromZone: Zone } | null>(null);

function onDragStart(card: { ID: string; Name: string }, fromUser: string, fromZone: string) {
  dragged.value = { card, fromUser, fromZone: fromZone as Zone };
}

function onDrop(toUser: string, toZone: string) {
  return (async () => {
    if (!dragged.value || !game.value) return;
    await moveCard({
      gameID: game.value.ID,
      user: toUser,
      fromUser: dragged.value.fromUser,
      cardID: dragged.value.card.ID,
      fromZone: dragged.value.fromZone,
      toZone: toZone as Zone,
    });
    dragged.value = null;
  })();
}

// Simple click-to-move: toggles between Hand and Battlefield for demo
async function quickMove(card: { ID: string; Name: string }, user: string, zone: string) {
  if (!game.value) return;
  const toZone: Zone = (zone === 'Hand' ? 'Battlefield' : 'Hand') as Zone;
  await moveCard({
    gameID: game.value.ID,
    user,
    fromUser: user,
    cardID: card.ID,
    fromZone: zone as Zone,
    toZone,
  });
}

type MoveCardArgs = {
  gameID: string;
  user: string;
  fromUser: string;
  cardID: string;
  fromZone: Zone;
  toZone: Zone;
};

async function moveCard(args: MoveCardArgs) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === args.user);
  if (!player || !player.Boardstate) return;

  // zones/type declared at module scope

  // Clone current zones
  const current: Record<Zone, { ID: string; Name: string }[]> = Object.fromEntries(
    zones.map(z => [z, [...(player.Boardstate?.[z as Zone] ?? [])]])
  ) as any;

  // Find full card details from source player's zones to preserve Name
  const sourcePlayer = g.Players.find(p => p.Username === args.fromUser);
  let movedCard: { ID: string; Name: string } | null = null;
  if (sourcePlayer?.Boardstate) {
    for (const z of zones) {
      const found = (sourcePlayer.Boardstate as any)[z]?.find((c: any) => c.ID === args.cardID);
      if (found) { movedCard = { ID: found.ID, Name: found.Name }; break; }
    }
  }

  // Remove from source zone (if same user)
  if (args.fromUser === args.user) {
    current[args.fromZone as Zone] = current[args.fromZone as Zone].filter(c => c.ID !== args.cardID);
  }
  // Add to destination zone (dedupe)
  if (!current[args.toZone as Zone].some(c => c.ID === args.cardID)) {
    current[args.toZone as Zone].push(movedCard ?? { ID: args.cardID, Name: '' });
  }

  const input: any = {
    UserID: player.ID ?? args.user,
    User: player.Username,
    GameID: g.ID,
    Life: player.Boardstate.Life,
    ...current,
  };

  try {
    await apolloClient.mutate({
      mutation: UPDATE_BOARDSTATE_MUTATION,
      variables: { input },
    });
    // Optimistically update local store
    applyLocalBoardstatePatch(args.user, draft => ({
      ...draft,
      ...current,
      Life: player.Boardstate!.Life,
    }));
    if (args.fromUser !== args.user) {
      // Remove card from source player locally (cross-player moves)
      applyLocalBoardstatePatch(args.fromUser, (draft: any) => ({
        ...draft,
        [args.fromZone]: (draft[args.fromZone] ?? []).filter((c: { ID: string }) => c.ID !== args.cardID),
      }));
    }
  } catch (e) {
    console.error('[board] moveCard failed', e);
  }
}

// Toolbar actions (self only)
async function draw(username: string) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === username);
  const top = player?.Boardstate?.Library?.[0];
  if (!top) return;
  await moveCard({
    gameID: g.ID,
    user: username,
    fromUser: username,
    cardID: top.ID,
    fromZone: 'Library',
    toZone: 'Hand',
  });
  addToast(`Drew ${top.Name}`);
}

async function mill(username: string) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === username);
  const top = player?.Boardstate?.Library?.[0];
  if (!top) return;
  await moveCard({
    gameID: g.ID,
    user: username,
    fromUser: username,
    cardID: top.ID,
    fromZone: 'Library',
    toZone: 'Graveyard',
  });
  addToast(`Milled ${top.Name}`);
}

async function revealTop(username: string) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === username);
  const top = player?.Boardstate?.Library?.[0];
  if (!top) return;
  await moveCard({
    gameID: g.ID,
    user: username,
    fromUser: username,
    cardID: top.ID,
    fromZone: 'Library',
    toZone: 'Revealed',
  });
  addToast(`Revealed ${top.Name}`);
}

async function shuffleLibrary(username: string) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === username);
  if (!player?.Boardstate?.Library || player.Boardstate.Library.length < 2) return;

  // Fisher-Yates shuffle (new array)
  const shuffled = [...player.Boardstate.Library];
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }

  const input: any = {
    UserID: player.ID ?? username,
    User: player.Username,
    GameID: g.ID,
    Life: player.Boardstate.Life,
    Commander: player.Boardstate.Commander ?? [],
    Battlefield: player.Boardstate.Battlefield ?? [],
    Hand: player.Boardstate.Hand ?? [],
    Graveyard: player.Boardstate.Graveyard ?? [],
    Exiled: player.Boardstate.Exiled ?? [],
    Revealed: player.Boardstate.Revealed ?? [],
    Controlled: player.Boardstate.Controlled ?? [],
    Library: shuffled,
  };

  try {
    await apolloClient.mutate({
      mutation: UPDATE_BOARDSTATE_MUTATION,
      variables: { input },
    });
    addToast('Shuffled library');
    // Optimistic local patch
    applyLocalBoardstatePatch(username, () => ({
      ...input,
    }));
  } catch (e) {
    console.error('[board] shuffleLibrary failed', e);
  }
}

async function changeLife(username: string, delta: number) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === username);
  if (!player?.Boardstate) return;

  const input: any = {
    UserID: player.ID ?? username,
    User: player.Username,
    GameID: g.ID,
    Life: (player.Boardstate.Life ?? 0) + delta,
    Commander: player.Boardstate.Commander ?? [],
    Battlefield: player.Boardstate.Battlefield ?? [],
    Hand: player.Boardstate.Hand ?? [],
    Graveyard: player.Boardstate.Graveyard ?? [],
    Exiled: player.Boardstate.Exiled ?? [],
    Revealed: player.Boardstate.Revealed ?? [],
    Controlled: player.Boardstate.Controlled ?? [],
    Library: player.Boardstate.Library ?? [],
  };

  try {
    await apolloClient.mutate({
      mutation: UPDATE_BOARDSTATE_MUTATION,
      variables: { input },
    });
    addToast(`${delta > 0 ? 'Gained' : 'Lost'} 1 life`);
    applyLocalBoardstatePatch(username, () => ({ ...input }));
  } catch (e) {
    console.error('[board] changeLife failed', e);
  }
}

// Scry 1 UX
const scry = ref<{ open: boolean; username: string; card?: { ID: string; Name: string } } | null>(null);

function scryOne(username: string) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === username);
  const top = player?.Boardstate?.Library?.[0];
  if (!top) return;
  scry.value = { open: true, username, card: { ID: top.ID, Name: top.Name } };
}

function scryKeepTop() {
  if (!scry.value) return;
  addToast(`Kept ${scry.value.card?.Name} on top`);
  scry.value = null;
}

async function scryPutBottom() {
  const g = game.value;
  const s = scry.value;
  if (!g || !s) return;
  const player = g.Players.find(p => p.Username === s.username);
  if (!player?.Boardstate?.Library || player.Boardstate.Library.length === 0) return;

  const [, ...rest] = player.Boardstate.Library;
  const newLibrary = [...rest, player.Boardstate.Library[0]];

  const input: any = {
    UserID: player.ID ?? s.username,
    User: player.Username,
    GameID: g.ID,
    Life: player.Boardstate.Life,
    Commander: player.Boardstate.Commander ?? [],
    Battlefield: player.Boardstate.Battlefield ?? [],
    Hand: player.Boardstate.Hand ?? [],
    Graveyard: player.Boardstate.Graveyard ?? [],
    Exiled: player.Boardstate.Exiled ?? [],
    Revealed: player.Boardstate.Revealed ?? [],
    Controlled: player.Boardstate.Controlled ?? [],
    Library: newLibrary,
  };

  try {
    await apolloClient.mutate({
      mutation: UPDATE_BOARDSTATE_MUTATION,
      variables: { input },
    });
    addToast(`Put ${s.card?.Name} on bottom`);
    applyLocalBoardstatePatch(s.username, () => ({ ...input }));
  } catch (e) {
    console.error('[board] scryPutBottom failed', e);
  } finally {
    scry.value = null;
  }
}

// Helper: patch a player's boardstate in the local active game
function applyLocalBoardstatePatch(username: string, updater: (prev: any) => any) {
  const root = games.activeGame as any;
  if (!root) return;
  const updatedPlayers = root.Players.map((p: any) => {
    if (p.Username !== username) return p;
    const prev = p.Boardstate ? { ...p.Boardstate } : {};
    const next = updater(prev);
    return { ...p, Boardstate: next };
  });
  // Replace the entire activeGame object to avoid mutating frozen Apollo results
  games.activeGame = { ...root, Players: updatedPlayers } as any;
}

// Toasts
type Toast = { id: number; text: string };
const toasts = ref<Toast[]>([]);
let toastCounter = 0;
function addToast(text: string, duration = 2500) {
  const id = ++toastCounter;
  toasts.value.push({ id, text });
  window.setTimeout(() => {
    toasts.value = toasts.value.filter(t => t.id !== id);
  }, duration);
}
</script>

<style scoped lang="scss">
.board {
  display: grid;
  gap: 1rem;
}

.board-header {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 16px;
  padding: 1rem 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.board-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: 1fr; /* vertical layout */
}

.players {
  display: grid;
  gap: 0.75rem;
}

.players article {
  background: rgba(255, 255, 255, 0.04);
  border-radius: 14px;
  padding: 0.75rem 1rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.players article.active {
  border-color: rgba(133, 215, 255, 0.6);
  box-shadow: 0 0 0 1px rgba(133, 215, 255, 0.15);
}

.zone {
  margin-top: 0.5rem;
}

.zone h3 {
  margin: 0 0 0.25rem;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: rgba(255, 255, 255, 0.65);
}

.zone h3 small {
  text-transform: none;
  letter-spacing: normal;
  font-weight: normal;
  font-size: 0.8em;
  color: rgba(255, 255, 255, 0.5);
}

.zone ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 0.25rem;
}

.card {
  padding: 0.2rem 0.45rem;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.08);
  font-size: 0.9rem;
  cursor: grab;
  user-select: none;
}

.card:active {
  cursor: grabbing;
}

.muted {
  opacity: 0.7;
}

.stack {
  background: rgba(255, 255, 255, 0.04);
  border-radius: 14px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  padding: 0.75rem 1rem;
  display: grid;
  gap: 0.5rem;
  position: sticky;
  top: 0.75rem;
  z-index: 5;
}

.stack ol {
  margin: 0;
  padding-left: 1.25rem;
  font-size: 0.9rem;
}

.loading-state {
  display: grid;
  place-items: center;
  min-height: 60vh;
  color: rgba(255, 255, 255, 0.6);
}

/* Toolbar */
header .player-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin-top: 0.4rem;
}

.player-toolbar .tool {
  appearance: none;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.06);
  color: #fff;
  font-size: 0.8rem;
  padding: 0.25rem 0.5rem;
  border-radius: 999px;
}

.player-toolbar .life-tools {
  display: inline-flex;
  gap: 0.3rem;
  margin-left: auto;
}

.player-toolbar .tool:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Scry modal */
.scry-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: grid;
  place-items: center;
}

.scry-modal {
  background: rgba(30, 30, 30, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 12px;
  padding: 1rem 1.25rem;
  min-width: 260px;
  max-width: 420px;
}

.scry-modal header {
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.scry-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.75rem;
}

/* Toasts */
.toasts {
  position: fixed;
  right: 1rem;
  bottom: 1rem;
  display: grid;
  gap: 0.5rem;
  z-index: 10;
}

.toast {
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: #fff;
  padding: 0.5rem 0.75rem;
  border-radius: 8px;
  font-size: 0.9rem;
}
</style>
