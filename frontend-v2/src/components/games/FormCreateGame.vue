<template>
  <div class="backdrop" @click.self="close">
    <section class="modal">
      <header>
        <h2>Start a new game</h2>
        <button class="link" @click="close" aria-label="Close">×</button>
      </header>
      <form @submit.prevent="handleSubmit">
        <label>
          <span>Game name</span>
          <input v-model="form.name" maxlength="64" placeholder="Friday Night Commander" />
        </label>
        <label>
          <span>Deck size</span>
          <input v-model.number="form.deckSize" type="number" min="60" max="250" />
        </label>
        <label>
          <span>Format</span>
          <select v-model="form.format">
            <option value="EDH">Commander</option>
          </select>
        </label>
        <label @keydown.stop>
          <span>Commander</span>
          <input
            v-model="commanderQuery"
            @input="onCommanderInput"
            @keydown.down.prevent="onCommanderKey('down')"
            @keydown.up.prevent="onCommanderKey('up')"
            @keydown.enter.prevent="onCommanderKey('enter')"
            @keydown.esc.prevent="onCommanderKey('escape')"
            @blur="onCommanderBlur"
            placeholder="Search for a commander (e.g., Atraxa)"
            autocomplete="off"
            role="combobox"
            :aria-expanded="showCommanderList ? 'true' : 'false'"
            aria-autocomplete="list"
            aria-controls="commander-typeahead"
            :aria-activedescendant="activeIndex >= 0 ? `commander-opt-${activeIndex}` : undefined"
          />
          <ul
            v-if="showCommanderList"
            id="commander-typeahead"
            class="typeahead"
            role="listbox"
          >
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
          <p v-if="selectedCommander" class="hint">
            Selected: {{ selectedCommander.Name }}
            <button type="button" class="link" @click="clearCommander">Clear</button>
          </p>
        </label>
        <label>
          <span>Decklist (CSV: quantity,name per line)</span>
          <textarea v-model="decklist" rows="6" placeholder="1, Atraxa, Pr…\n99, Basic Island"></textarea>
          <p class="hint">Deck count: {{ deckCount }}</p>
        </label>
        <footer>
          <button type="button" class="secondary" @click="close">Cancel</button>
          <button type="submit" class="primary" :disabled="games.loading">
            {{ games.loading ? 'Creating…' : 'Create game' }}
          </button>
        </footer>
      </form>
    </section>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed } from 'vue';
import { useGamesStore } from '../../stores/games';
import { useAuthStore } from '../../stores/auth';
import { apolloClient } from '../../services/apollo';
import { SEARCH_CARDS_QUERY } from '../../graphql/queries';

const emit = defineEmits<{ (event: 'close'): void; (event: 'created', id: string): void }>();

const games = useGamesStore();
const auth = useAuthStore();

const form = reactive({
  name: '',
  deckSize: 99,
  format: 'EDH',
});

// Commander search state
const commanderQuery = ref('');
const commanderResults = ref<{ ID: string; Name: string }[]>([]);
const selectedCommander = ref<{ ID: string; Name: string } | null>(null);
const showCommanderList = ref(false);
const isSearching = ref(false);
const activeIndex = ref(-1);
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
    const { data } = await apolloClient.query<{ search: { ID: string; Name: string }[] }>({
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
  // Delay to allow click selection to register before closing.
  setTimeout(() => {
    showCommanderList.value = false;
  }, 120);
}

function selectCommander(card: { ID: string; Name: string }) {
  selectedCommander.value = card;
  commanderQuery.value = card.Name;
  showCommanderList.value = false;
}

function clearCommander() {
  selectedCommander.value = null;
}

// Decklist raw CSV input
const decklist = ref('');
const deckCount = computed(() => {
  if (!decklist.value) return 0;
  let count = 0;
  for (const line of decklist.value.split(/\r?\n/)) {
    const trimmed = line.trim();
    if (!trimmed) continue;
    const [qtyRaw] = trimmed.split(',');
    const qty = parseInt(qtyRaw, 10);
    if (!Number.isNaN(qty)) count += qty;
    else count += 1;
  }
  return count;
});

function close() {
  emit('close');
}

async function handleSubmit() {
  const newId = crypto.randomUUID();
  const payload = {
    ID: newId,
    Turn: { Player: auth.profile?.Username ?? 'Unknown', Phase: 'MAIN', Number: 1 },
    Players: [
      {
        UserID: auth.profile?.ID ?? '',
        User: auth.profile?.Username ?? '',
        GameID: newId,
        Life: 40,
        Decklist: decklist.value,
        Commander: selectedCommander.value ? [{ ID: selectedCommander.value.ID, Name: selectedCommander.value.Name }] : [],
        Library: [],
        Graveyard: [],
        Exiled: [],
        Battlefield: [],
        Hand: [],
        Revealed: [],
        Controlled: [],
        Counters: [],
      },
    ],
  } as const;
  const gameID = await games.createGame(payload);
  if (gameID) {
    emit('created', gameID);
  }
}
</script>

<style scoped lang="scss">
.backdrop {
  position: fixed;
  inset: 0;
  background: rgba(6, 8, 11, 0.7);
  display: grid;
  place-items: center;
  z-index: 40;
}

.modal {
  width: min(90vw, 420px);
  background: rgba(18, 21, 28, 0.95);
  border-radius: 18px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  padding: 1.75rem;
  display: grid;
  gap: 1.25rem;
}

header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

form {
  display: grid;
  gap: 1rem;
}

label {
  display: grid;
  gap: 0.5rem;
}

input,
select {
  padding: 0.7rem 0.9rem;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.05);
  color: inherit;
}
footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

button {
  border: none;
  border-radius: 8px;
  padding: 0.6rem 1.2rem;
  font-weight: 600;
  cursor: pointer;
}

button.primary {
  background: linear-gradient(120deg, #85d7ff, #26c2ff);
  color: #0b1016;
}

button.secondary,
button.link {
  background: transparent;
  color: rgba(255, 255, 255, 0.78);
}

button.link {
  font-size: 1.5rem;
  padding: 0;
  line-height: 1;
}

.typeahead {
  margin-top: 0.25rem;
  list-style: none;
  padding: 0;
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  max-height: 200px;
  overflow: auto;
}
.typeahead li {
  padding: 0.4rem 0.6rem;
  cursor: pointer;
}
.typeahead li:hover {
  background: rgba(255,255,255,0.06);
}
.typeahead li.active {
  background: rgba(255,255,255,0.12);
}
.hint {
  opacity: 0.8;
  font-size: 0.9rem;
}
</style>
