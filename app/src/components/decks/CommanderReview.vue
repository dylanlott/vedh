<template>
  <section
    class="commander-review"
    :data-valid-selection="selectionValid"
    aria-labelledby="commander-review-heading"
  >
    <header>
      <h2 id="commander-review-heading">Choose your commander</h2>
    </header>

    <div v-if="candidates.length === 0" class="empty-state">
      <h3>No legal commander detected yet</h3>
      <p>None of your parsed cards can legally be a Commander. Search for one manually, or check your list for a commander-eligible creature or planeswalker.</p>
    </div>

    <div
      v-else
      class="candidate-list typeahead"
      data-testid="commander-candidates"
      role="listbox"
      aria-label="Detected commander candidates"
      aria-multiselectable="true"
    >
      <button
        v-for="candidate in candidates.slice(0, 3)"
        :key="candidate.ID"
        type="button"
        class="candidate"
        :class="{ selected: isSelected(candidate) }"
        data-testid="commander-candidate"
        :aria-pressed="isSelected(candidate)"
        @click="toggleCandidate(candidate)"
      >
        <span class="candidate-name" data-testid="candidate-name" :title="candidate.Name">
          {{ truncateDisplay(candidate.Name) }}
        </span>
      </button>
    </div>

    <div v-if="localSelected.length" class="selected-commanders chips" aria-label="Selected commanders">
      <span v-for="(commander, index) in localSelected" :key="commander.ID" class="selected-chip">
        <span class="truncated" :title="commander.Name">{{ truncateDisplay(commander.Name) }}</span>
        <button
          type="button"
          class="icon-control"
          :aria-label="`Remove ${commander.Name} from commander selection`"
          @click="removeCommander(index)"
        >
          ×
        </button>
      </span>
    </div>

    <p v-if="constraintError" class="constraint" data-testid="partner-constraint">
      {{ constraintError }}
    </p>

    <label class="search-all">
      <span>Search all cards</span>
      <input
        v-model="searchQuery"
        data-testid="commander-search"
        type="text"
        autocomplete="off"
        role="combobox"
        aria-controls="commander-search-results"
        :aria-expanded="showSearchResults"
        @input="onSearchInput"
        @keydown.down.prevent="moveActive(1)"
        @keydown.up.prevent="moveActive(-1)"
        @keydown.enter.prevent="chooseActive"
        @keydown.esc.prevent="closeSearch"
      />
      <ul
        v-if="showSearchResults"
        id="commander-search-results"
        class="typeahead search-results"
        role="listbox"
      >
        <li v-if="searching" class="search-hint" role="option" aria-disabled="true">Searching…</li>
        <template v-else>
          <li
            v-for="(card, index) in searchResults"
            :key="card.ID"
            :class="{ active: index === activeIndex }"
            role="option"
            :aria-selected="index === activeIndex"
            @mousedown.prevent="toggleCandidate(card)"
            @mousemove="activeIndex = index"
          >
            <button type="button" data-testid="manual-commander-result" :title="card.Name">
              {{ truncateDisplay(card.Name) }}
            </button>
          </li>
          <li v-if="searchResults.length === 0" class="search-hint" role="option" aria-disabled="true">No results</li>
        </template>
      </ul>
    </label>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { apolloClient } from '../../services/apollo';
import { SEARCH_CARDS_QUERY } from '../../graphql/queries';
import {
  canAddSecondCommander,
  isValidPartnerPair,
  partnerConstraintMessage,
  type CommanderPick,
} from '../../services/commanderPartner';
import type { Card } from '../../types/generated';

const props = defineProps<{
  candidates: CommanderPick[];
  selected: CommanderPick[];
}>();

const emit = defineEmits<{
  (event: 'selection-change', selected: CommanderPick[]): void;
}>();

const localSelected = ref<CommanderPick[]>([...props.selected]);
const constraintError = ref('');
const searchQuery = ref('');
const searchResults = ref<CommanderPick[]>([]);
const searching = ref(false);
const showSearchResults = ref(false);
const activeIndex = ref(-1);
let searchDebounce: number | undefined;

watch(() => props.selected, (selected) => {
  localSelected.value = [...selected];
}, { deep: true });

const selectionValid = computed(() => {
  if (constraintError.value || localSelected.value.length === 0) return false;
  if (localSelected.value.length === 1) return true;
  return localSelected.value.length === 2
    && isValidPartnerPair(localSelected.value[0], localSelected.value[1]);
});

function isSelected(card: CommanderPick): boolean {
  return localSelected.value.some((selected) => selected.ID === card.ID);
}

function publish(): void {
  emit('selection-change', [...localSelected.value]);
}

function toggleCandidate(card: CommanderPick | Card): void {
  const candidate = card as CommanderPick;
  const existing = localSelected.value.findIndex((selected) => selected.ID === candidate.ID);
  if (existing >= 0) {
    localSelected.value.splice(existing, 1);
    constraintError.value = '';
    publish();
    return;
  }
  if (localSelected.value.length === 0) {
    localSelected.value.push(candidate);
    constraintError.value = '';
    publish();
    closeSearch();
    return;
  }
  if (localSelected.value.length >= 2) return;

  const first = localSelected.value[0];
  if (!canAddSecondCommander(localSelected.value) || !isValidPartnerPair(first, candidate)) {
    constraintError.value = partnerConstraintMessage(first);
    return;
  }

  localSelected.value.push(candidate);
  constraintError.value = '';
  publish();
  closeSearch();
}

function removeCommander(index: number): void {
  localSelected.value.splice(index, 1);
  constraintError.value = '';
  publish();
}

function onSearchInput(): void {
  const query = searchQuery.value.trim();
  showSearchResults.value = query.length >= 2;
  activeIndex.value = -1;
  if (searchDebounce !== undefined) window.clearTimeout(searchDebounce);
  if (!showSearchResults.value) {
    searchResults.value = [];
    searching.value = false;
    return;
  }
  searchDebounce = window.setTimeout(() => void runSearch(query), 150);
}

async function runSearch(query: string): Promise<void> {
  searching.value = true;
  try {
    const { data } = await apolloClient.query<{ search?: CommanderPick[] }>({
      query: SEARCH_CARDS_QUERY,
      variables: { name: `%${query}%` },
      fetchPolicy: 'no-cache',
    });
    searchResults.value = (data?.search ?? []).slice(0, 8);
  } catch {
    searchResults.value = [];
  } finally {
    searching.value = false;
  }
}

function moveActive(delta: number): void {
  if (!showSearchResults.value || searchResults.value.length === 0) return;
  const count = searchResults.value.length;
  activeIndex.value = (activeIndex.value + delta + count) % count;
}

function chooseActive(): void {
  const active = searchResults.value[activeIndex.value];
  if (active) toggleCandidate(active);
}

function closeSearch(): void {
  showSearchResults.value = false;
  searchQuery.value = '';
  activeIndex.value = -1;
}

function truncateDisplay(value: string, maximum = 48): string {
  const characters = Array.from(value);
  return characters.length <= maximum ? value : `${characters.slice(0, maximum).join('')}…`;
}

onBeforeUnmount(() => {
  if (searchDebounce !== undefined) window.clearTimeout(searchDebounce);
});
</script>

<style scoped lang="scss">
@use '@/styles/breakpoints' as bp;

.commander-review {
  min-width: 0;
  display: grid;
  gap: 16px;
  padding: 24px;
  border: 1px solid var(--vedh-border);
  border-radius: 20px;
  background: var(--vedh-panel);
}

header,
h2,
h3,
p {
  margin: 0;
}

h2,
h3 {
  font-size: 20px;
  font-weight: 600;
  line-height: 1.2;
}

.empty-state,
.search-all {
  display: grid;
  gap: 8px;
}

.candidate-list,
.selected-commanders {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.typeahead {
  max-height: 200px;
  overflow: auto;
}

.candidate {
  min-width: 0;
  padding: 12px 16px;
  border: 1px solid var(--vedh-border);
  border-radius: 10px;
  background: rgba(255, 244, 237, 0.05);
  color: var(--vedh-text);
  text-align: left;
  cursor: pointer;
}

.candidate.selected {
  border-color: var(--vedh-primary);
  box-shadow: inset 0 0 0 1px var(--vedh-primary);
}

.candidate-name,
.truncated {
  display: block;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chips {
  flex-direction: row;
  flex-wrap: wrap;
}

.selected-chip {
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding-left: 8px;
  border: 1px solid var(--vedh-border);
  border-radius: 999px;
  background: var(--vedh-surface);
}

.icon-control {
  flex: 0 0 44px;
  min-width: 44px;
  min-height: 44px;
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--vedh-text);
  cursor: pointer;
}

.constraint {
  color: var(--vedh-danger);
}

input {
  box-sizing: border-box;
  width: 100%;
  min-width: 0;
  padding: 12px 16px;
  border: 1px solid var(--vedh-border);
  border-radius: 10px;
  background: rgba(255, 244, 237, 0.05);
  color: var(--vedh-text);
  font: inherit;
}

.search-results {
  margin: 0;
  padding: 0;
  list-style: none;
  border: 1px solid var(--vedh-border);
}

.search-results li,
.search-results button {
  box-sizing: border-box;
  width: 100%;
}

.search-results button {
  padding: 8px 12px;
  border: 0;
  background: transparent;
  color: var(--vedh-text);
  text-align: left;
}

.search-results li.active {
  background: var(--vedh-surface-strong);
}

@media (min-width: bp.$breakpoint-tablet) {
  .candidate-list {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>
