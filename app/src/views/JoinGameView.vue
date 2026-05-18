<template>
  <section class="join-game">
    <h1>Join game</h1>
    
    <form v-if="!gameID" @submit.prevent="submitInvite">
      <label for="invite">Paste invite link or game ID</label>
      <input id="invite" v-model.trim="invite" type="text" placeholder="e.g. https://app/join/abcd-1234 or abcd-1234" />
      <div class="help" v-if="invite && !parsedID">Couldn’t detect a game ID from that input.</div>
      <button class="primary" :disabled="!parsedID">Continue</button>
    </form>

    <form v-else @submit.prevent="handleJoin">
      <p>You are about to join game <strong>{{ gameID }}</strong>.</p>

      <label class="stacked">
        <span>Commander(s) — up to 2 (Partners)</span>
        <div class="inline">
          <button class="secondary" type="button" @click="isCommanderModalOpen = true">Choose commander</button>
          <div class="chips" v-if="selectedCommanders.length">
            <span v-for="(cmd, idx) in selectedCommanders" :key="cmd.ID" class="chip">
              {{ cmd.Name }}
              <button type="button" class="remove" @click="removeCommander(idx)" aria-label="Remove">×</button>
            </span>
            <button type="button" class="link" @click="clearAllCommanders">Clear all</button>
          </div>
          <span class="hint" v-else>No commander selected</span>
        </div>
      </label>

      <label class="stacked">
        <span>Decklist (CSV: quantity,name per line)</span>
        <textarea v-model="decklist" rows="6" placeholder="1, Atraxa, Pr…\n99, Basic Island"></textarea>
        <p class="hint">Deck count: {{ deckCount }}</p>
      </label>

      <footer class="actions">
        <button class="primary" :disabled="games.loading">{{ games.loading ? 'Joining…' : 'Join game' }}</button>
      </footer>
    </form>

    <!-- Commander selection modal -->
    <div v-if="isCommanderModalOpen" class="backdrop" @click.self="isCommanderModalOpen = false">
      <section class="modal">
        <header>
          <h2>Select your Commander</h2>
          <button class="link" type="button" @click="isCommanderModalOpen = false" aria-label="Close">×</button>
        </header>
        <label @keydown.stop>
          <span>Search</span>
          <input
            v-model="commanderQuery"
            @input="onCommanderInput"
            @keydown.down.prevent="onCommanderKey('down')"
            @keydown.up.prevent="onCommanderKey('up')"
            @keydown.enter.prevent="onCommanderKey('enter')"
            @keydown.esc.prevent="onCommanderKey('escape')"
            @blur="onCommanderBlur"
            placeholder="e.g., Atraxa"
            autocomplete="off"
            role="combobox"
            :aria-expanded="showCommanderList ? 'true' : 'false'"
            aria-autocomplete="list"
            aria-controls="commander-typeahead"
            :aria-activedescendant="activeIndex >= 0 ? `commander-opt-${activeIndex}` : undefined"
          />
          <ul v-if="showCommanderList" id="commander-typeahead" class="typeahead" role="listbox">
            <li v-if="isSearching" class="hint" role="option" aria-disabled="true">Searching…</li>
            <template v-else>
              <li
                v-for="(c, idx) in limitedCommanderResults"
                :key="c.ID"
                :id="`commander-opt-${idx}`"
                role="option"
                :aria-selected="idx === activeIndex ? 'true' : 'false'"
                :class="{ active: idx === activeIndex }"
                @mousedown.prevent="selectCommander(c)"
                @mousemove="activeIndex = idx"
              >
                {{ c.Name }}
              </li>
              <li v-if="!limitedCommanderResults.length" class="hint" role="option" aria-disabled="true">No results</li>
            </template>
          </ul>
          <div class="chips" v-if="selectedCommanders.length">
            <span v-for="(cmd, idx) in selectedCommanders" :key="cmd.ID" class="chip">
              {{ cmd.Name }}
              <button type="button" class="remove" @click="removeCommander(idx)" aria-label="Remove">×</button>
            </span>
            <button type="button" class="link" @click="clearAllCommanders">Clear all</button>
          </div>
          <p v-if="commanderError" class="hint">{{ commanderError }}</p>
          <p v-else-if="!selectedCommanders.length" class="hint">No commander selected</p>
        </label>
        <footer class="actions">
          <button class="secondary" type="button" @click="isCommanderModalOpen = false">Done</button>
        </footer>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useGamesStore } from '../stores/games';
import { useAuthStore } from '../stores/auth';
import { apolloClient } from '../services/apollo';
import { SEARCH_CARDS_QUERY } from '../graphql/queries';
import {
  type CommanderPick,
  canAddSecondCommander,
  isValidPartnerPair,
  partnerConstraintMessage,
} from '../services/commanderPartner';

const route = useRoute();
const router = useRouter();
const games = useGamesStore();
const auth = useAuthStore();

const gameID = computed(() => route.params.id as string | undefined);
const invite = ref('');

const parsedID = computed(() => {
  const raw = invite.value.trim();
  if (!raw) return '';
  // Match URLs containing /join/:id or /games/:id
  const match = raw.match(/\/(?:join|games)\/([A-Za-z0-9\-]+)/i);
  if (match?.[1]) return match[1];
  // If it looks like a bare id (uuid-ish or slug), accept alnum-dash 6+ chars
  if (/^[A-Za-z0-9\-]{6,}$/.test(raw)) return raw;
  return '';
});

function submitInvite() {
  if (!parsedID.value) return;
  router.push({ name: 'join-game', params: { id: parsedID.value } });
}

// Commander search state (mirrors FormCreateGame)
const isCommanderModalOpen = ref(false);
const commanderQuery = ref('');
const commanderResults = ref<CommanderPick[]>([]);
const selectedCommanders = ref<CommanderPick[]>([]);
const showCommanderList = ref(false);
const isSearching = ref(false);
const activeIndex = ref(-1);
const commanderError = ref<string>('');
const resultsLimit = 8;
let commanderDebounce: number | undefined;

const limitedCommanderResults = computed(() => commanderResults.value.slice(0, resultsLimit));

async function runCommanderSearch(query: string) {
  if (query.length < 2) {
    commanderResults.value = [];
    isSearching.value = false;
    return;
  }
  isSearching.value = true;
  try {
    const { data } = await apolloClient.query<{ search?: CommanderPick[] }>({
      query: SEARCH_CARDS_QUERY,
      variables: { name: `%${query}%` },
      fetchPolicy: 'no-cache',
    });
    commanderResults.value = data?.search ?? [];
  } catch (e) {
    commanderResults.value = [];
  } finally {
    isSearching.value = false;
  }
}

function onCommanderInput() {
  showCommanderList.value = commanderQuery.value.length >= 2;
  activeIndex.value = -1;
  if (commanderDebounce) window.clearTimeout(commanderDebounce);
  commanderDebounce = window.setTimeout(() => {
    void runCommanderSearch(commanderQuery.value.trim());
  }, 150);
}

function onCommanderKey(key: 'down' | 'up' | 'enter' | 'escape') {
  if (!showCommanderList.value) {
    if (key === 'down') {
      showCommanderList.value = commanderQuery.value.length >= 2;
      if (showCommanderList.value && !limitedCommanderResults.value.length) void runCommanderSearch(commanderQuery.value.trim());
    }
    return;
  }
  const max = limitedCommanderResults.value.length - 1;
  if (key === 'down') {
    activeIndex.value = activeIndex.value < max ? activeIndex.value + 1 : 0;
  } else if (key === 'up') {
    activeIndex.value = activeIndex.value > 0 ? activeIndex.value - 1 : max;
  } else if (key === 'enter') {
    if (activeIndex.value >= 0 && activeIndex.value <= max) {
      selectCommander(limitedCommanderResults.value[activeIndex.value]);
    }
  } else if (key === 'escape') {
    showCommanderList.value = false;
  }
}

function onCommanderBlur() {
  setTimeout(() => {
    showCommanderList.value = false;
  }, 120);
}

function selectCommander(card: { ID: string; Name: string }) {
  commanderError.value = '';
  const exists = selectedCommanders.value.some(c => c.ID === card.ID);

  if (exists) {
    commanderQuery.value = '';
    showCommanderList.value = false;
    return;
  }

  if (selectedCommanders.value.length === 0) {
    selectedCommanders.value.push(card as CommanderPick);
  } else if (selectedCommanders.value.length === 1) {
    const first = selectedCommanders.value[0];
    if (!canAddSecondCommander(selectedCommanders.value)) {
      commanderError.value = partnerConstraintMessage(first);
    } else if (!isValidPartnerPair(first, card as CommanderPick)) {
      commanderError.value = partnerConstraintMessage(first);
    } else {
      selectedCommanders.value.push(card as CommanderPick);
    }
  }
  commanderQuery.value = '';
  showCommanderList.value = false;
}

function removeCommander(index: number) {
  selectedCommanders.value.splice(index, 1);
  commanderError.value = '';
}

function clearAllCommanders() {
  selectedCommanders.value = [];
  commanderError.value = '';
}

// Decklist
const decklist = ref('');
const deckCount = computed(() => {
  if (!decklist.value) return 0;
  let count = 0;
  for (const line of decklist.value.split(/\r?\n/)) {
    const trimmed = line.trim();
    if (!trimmed) continue;
    const [qtyRaw] = trimmed.split(',');
    const qty = parseInt(qtyRaw, 10);
    if (!Number.isNaN(qty)) count += qty; else count += 1;
  }
  return count;
});

async function handleJoin() {
  if (!gameID.value || !auth.profile) return;
  const payload = {
    ID: gameID.value,
    Decklist: decklist.value,
    BoardState: {
      UserID: auth.profile.ID,
      User: auth.profile.Username,
      GameID: gameID.value,
      Life: 40,
  Commander: selectedCommanders.value.map(c => ({ ID: c.ID, Name: c.Name })),
      Library: [],
      Graveyard: [],
      Exiled: [],
      Battlefield: [],
      Hand: [],
      Revealed: [],
      Controlled: [],
      Counters: [],
    },
  } as const;
  const joinedID = await games.joinGame(payload);
  if (joinedID) {
    router.push({ name: 'board', params: { id: joinedID } });
  }
}
</script>

<style scoped lang="scss">
.join-game {
  max-width: 480px;
  margin: 3rem auto;
  padding: 2.5rem;
  border-radius: 20px;
  background: var(--vedh-panel);
  border: 1px solid var(--vedh-border);
}

button.primary {
  border: none;
  border-radius: 10px;
  padding: 0.75rem 1rem;
  font-size: 1rem;
  font-weight: 600;
  background: var(--vedh-primary-gradient);
  color: var(--vedh-primary-contrast);
  cursor: pointer;
}

button.secondary {
  border: 1px solid var(--vedh-border);
  border-radius: 10px;
  padding: 0.6rem 0.9rem;
  background: rgba(255, 244, 237, 0.05);
  color: var(--vedh-text);
  cursor: pointer;
}

.stacked { display: grid; gap: 0.5rem; }
.inline { display: flex; gap: 0.75rem; align-items: center; }
.hint { opacity: 0.8; font-size: 0.9rem; }

.backdrop {
  position: fixed;
  inset: 0;
  background: rgba(33, 20, 18, 0.72);
  display: grid;
  place-items: center;
  z-index: 40;
}

.modal {
  width: min(90vw, 420px);
  background: var(--vedh-panel-strong);
  border-radius: 18px;
  border: 1px solid var(--vedh-border);
  padding: 1.5rem;
  display: grid;
  gap: 0.9rem;
}

header { display: flex; align-items: center; justify-content: space-between; }
label { display: grid; gap: 0.5rem; }
input, textarea { 
  padding: 0.7rem 0.9rem; 
  border-radius: 10px; 
  border: 1px solid var(--vedh-border);
  background: rgba(255,244,237,0.05);
  color: inherit;
}
.actions { display: flex; justify-content: flex-end; gap: 0.75rem; }

.typeahead {
  margin-top: 0.25rem;
  list-style: none;
  padding: 0;
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  max-height: 200px;
  overflow: auto;
}
.typeahead li { padding: 0.4rem 0.6rem; cursor: pointer; }
.typeahead li:hover { background: rgba(255,255,255,0.06); }
.typeahead li.active { background: rgba(255,255,255,0.12); }

.chips { display: flex; flex-wrap: wrap; gap: 0.4rem; align-items: center; }
.chip {
  display: inline-flex;
  gap: 0.4rem;
  align-items: center;
  padding: 0.25rem 0.5rem;
  border-radius: 999px;
  background: rgba(255,255,255,0.08);
  border: 1px solid rgba(255,255,255,0.12);
  font-size: 0.9rem;
}
.chip .remove {
  background: transparent;
  border: none;
  color: inherit;
  cursor: pointer;
  padding: 0 0.25rem;
}
</style>
