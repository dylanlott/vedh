<template>
  <section class="board" v-if="game">
    <header class="board-header">
      <h1>Game {{ game.ID }}</h1>
      <p>Turn {{ game.Turn?.Number ?? '—' }} • {{ game.Turn?.Player ?? 'Unknown player' }} ({{ game.Turn?.Phase ?? 'Phase' }})</p>
    </header>

    <div class="board-grid">
      <aside class="players">
        <article v-for="player in game.Players" :key="player.ID" :class="{ active: isActivePlayer(player.Username) }">
          <header>
            <h2>{{ player.Username }}</h2>
            <span class="life">{{ player.Boardstate?.Life ?? '—' }} life</span>
          </header>
          <div class="zone" :data-zone="'Commander'" @dragover.prevent @drop="onDrop(player.Username, 'Commander')">
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
          <div class="zone" :data-zone="'Battlefield'" @dragover.prevent @drop="onDrop(player.Username, 'Battlefield')">
            <h3>Battlefield</h3>
            <ul>
                <li v-for="card in player.Boardstate?.Battlefield ?? []" :key="card.ID" class="card" draggable="true" @dragstart="onDragStart(card, player.Username, 'Battlefield')" @click="quickMove(card, player.Username, 'Battlefield')">
                  {{ card.Name }}
                </li>
            </ul>
          </div>
          <div class="zone" :data-zone="'Hand'" @dragover.prevent @drop="onDrop(player.Username, 'Hand')">
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
        </article>
      </aside>

      <section class="stack">
        <header>
          <h2>Stack</h2>
        </header>
        <ol>
          <li v-for="card in game.Stack" :key="card.ID" class="card">{{ card.Name }}</li>
        </ol>
      </section>
    </div>
  </section>
  <section v-else class="loading-state">
    <p>Loading game…</p>
  </section>
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

// Basic drag-and-drop state
const dragged = ref<{ card: { ID: string; Name: string }; fromUser: string; fromZone: string } | null>(null);

function onDragStart(card: { ID: string; Name: string }, fromUser: string, fromZone: string) {
  dragged.value = { card, fromUser, fromZone };
}

function onDrop(toUser: string, toZone: string) {
  return async () => {
    if (!dragged.value || !game.value) return;
    await moveCard({
      gameID: game.value.ID,
      user: toUser,
      fromUser: dragged.value.fromUser,
      cardID: dragged.value.card.ID,
      fromZone: dragged.value.fromZone,
      toZone,
    });
    dragged.value = null;
  };
}

// Simple click-to-move: toggles between Hand and Battlefield for demo
async function quickMove(card: { ID: string; Name: string }, user: string, zone: string) {
  if (!game.value) return;
  const toZone = zone === 'Hand' ? 'Battlefield' : 'Hand';
  await moveCard({
    gameID: game.value.ID,
    user,
    fromUser: user,
    cardID: card.ID,
    fromZone: zone,
    toZone,
  });
}

type MoveCardArgs = {
  gameID: string;
  user: string;
  fromUser: string;
  cardID: string;
  fromZone: string;
  toZone: string;
};

async function moveCard(args: MoveCardArgs) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === args.user);
  if (!player || !player.Boardstate) return;

  const zones = ['Commander','Battlefield','Hand','Graveyard','Exiled','Revealed','Library','Controlled'] as const;
  type Zone = typeof zones[number];

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
  } catch (e) {
    console.error('[board] moveCard failed', e);
  }
}
</script>

<style scoped lang="scss">
.board {
  display: grid;
  gap: 1.5rem;
}

.board-header {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 16px;
  padding: 1.5rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.board-grid {
  display: grid;
  gap: 1.5rem;
  grid-template-columns: minmax(260px, 320px) 1fr;
}

.players {
  display: grid;
  gap: 1rem;
}

.players article {
  background: rgba(255, 255, 255, 0.04);
  border-radius: 14px;
  padding: 1rem 1.2rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.players article.active {
  border-color: rgba(133, 215, 255, 0.6);
  box-shadow: 0 0 0 1px rgba(133, 215, 255, 0.15);
}

.zone {
  margin-top: 0.75rem;
}

.zone h3 {
  margin: 0 0 0.35rem;
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: rgba(255, 255, 255, 0.65);
}

.zone ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 0.35rem;
}

.card {
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.08);
  cursor: grab;
  user-select: none;
}

.card:active {
  cursor: grabbing;
}

.stack {
  background: rgba(255, 255, 255, 0.04);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 1.5rem;
  display: grid;
  gap: 1rem;
}

.stack ol {
  margin: 0;
  padding-left: 1.25rem;
}

.loading-state {
  display: grid;
  place-items: center;
  min-height: 60vh;
  color: rgba(255, 255, 255, 0.6);
}
</style>
