<template>
  <section class="defaults" aria-labelledby="default-deck-heading">
    <div><strong id="default-deck-heading">Start with a default deck</strong><p class="hint">Curated snapshots across all five Commander Brackets. Applying one replaces the commander and decklist below.</p></div>
    <div class="picker-row">
      <select v-model="selectedID" aria-label="Default deck">
        <option value="">Choose a bracketed deck…</option>
        <option v-for="deck in defaultDecks" :key="deck.id" :value="deck.id">B{{ deck.bracket }} · {{ deck.bracketName }} — {{ deck.name }}</option>
      </select>
      <button type="button" class="secondary" :disabled="!selectedDeck" @click="apply">Apply deck</button>
    </div>
    <div v-if="selectedDeck" class="deck-summary"><strong>{{ selectedDeck.commander }}</strong><span>{{ selectedDeck.description }}</span><small>{{ deckCardCount(selectedDeck) }} cards · Updated {{ selectedDeck.updatedAt }}</small><a :href="selectedDeck.sourceUrl" target="_blank" rel="noreferrer">{{ selectedDeck.sourceLabel }}</a></div>
  </section>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue';
import { defaultDecks, deckCardCount, type DefaultDeck } from '../../decks/defaultDecks';
const emit = defineEmits<{ (event: 'apply', deck: DefaultDeck): void }>();
const selectedID = ref('');
const selectedDeck = computed(() => defaultDecks.find(deck => deck.id === selectedID.value));
function apply() { if (selectedDeck.value) emit('apply', selectedDeck.value); }
</script>
<style scoped lang="scss">
.defaults { display: grid; gap: .65rem; padding: .8rem; border: 1px solid var(--vedh-border); border-radius: 10px; background: rgba(255,255,255,.03); }
.defaults p { margin: .2rem 0 0; }.picker-row { display: flex; gap: .5rem; align-items: center; }.picker-row select { min-width: 0; flex: 1; }.picker-row button { white-space: nowrap; }.picker-row button:disabled { cursor: not-allowed; opacity: .5; }.deck-summary { display: grid; gap: .2rem; font-size: .9rem; }.deck-summary span,.deck-summary small { opacity: .8; }.deck-summary a { color: inherit; width: fit-content; font-size: .82rem; }@media(max-width:520px){.picker-row{align-items:stretch;flex-direction:column}}
</style>
