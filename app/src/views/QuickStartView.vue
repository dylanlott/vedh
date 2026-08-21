<template>
  <section class="quick-start">
    <h1 class="display-heading">Paste a deck. Start a table.</h1>
    <MobileHeadsUpNotice />

    <div class="quick-start-layout">
      <DeckImportPanel
        :class="{ 'quick-start-deck-import--wide': stage === 'import' }"
        :initial-text="deckText"
        :initial-source-u-r-l="sourceURL"
        :session-id="sessionID"
        persistence-key="edhgo/quickstart-draft"
        @submit="handleDeckImportStarted"
        @change="handleDeckChange"
        @preview-resolved="handlePreviewResolved"
        @continue="stage = 'commander'"
        @error="handlePanelError"
      />

      <section v-if="stage === 'commander' || selectedCommanders.length" class="finish-pane">
        <CommanderReview
          :candidates="commanderCandidates"
          :selected="selectedCommanders"
          @selection-change="handleCommanderChange"
        />

        <label class="display-name-field">
          <span>Display name (optional)</span>
          <input
            v-model="displayName"
            data-testid="display-name"
            type="text"
            maxlength="64"
            placeholder="Optional — skip it and we'll give you a name like 'Brave Sliver'."
          />
        </label>

        <footer class="actions">
          <button
            class="primary"
            type="button"
            data-testid="start-table"
            :disabled="!canStartTable"
            @click="handleStartTable"
          >
            {{ creating ? 'Starting your table…' : 'Start my table' }}
          </button>
        </footer>
      </section>
    </div>

    <section v-if="pageError" class="page-error" data-testid="quick-start-error" role="alert">
      <h2>{{ pageError.title }}</h2>
      <p>{{ pageError.body }}</p>
      <button class="secondary" type="button" @click="handleStartTable">{{ pageError.affordance }}</button>
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import CommanderReview from '../components/decks/CommanderReview.vue';
import DeckImportPanel, { type DeckImportChange } from '../components/decks/DeckImportPanel.vue';
import MobileHeadsUpNotice from '../components/decks/MobileHeadsUpNotice.vue';
import {
  resolveActivationError,
  type ActivationErrorEntry,
} from '../services/activationErrors';
import type { CommanderPick } from '../services/commanderPartner';
import { getSessionID, track } from '../services/productEvents';
import {
  clearDraft,
  readDraft,
  saveDraft,
  type QuickStartDraft,
} from '../services/quickStartDraft';
import { useAuthStore } from '../stores/auth';
import { useGamesStore } from '../stores/games';
import type { DeckPreview } from '../types/generated';

const router = useRouter();
const auth = useAuthStore();
const games = useGamesStore();
const sessionID = getSessionID();
const storedDraft = readDraft();

const deckText = ref(storedDraft?.deckText ?? '');
const sourceURL = ref(storedDraft?.sourceURL ?? '');
const corrections = ref<Record<number, string>>({ ...(storedDraft?.corrections ?? {}) });
const selectedCommanders = ref<CommanderPick[]>([...(storedDraft?.selectedCommanders ?? [])]);
const displayName = ref(storedDraft?.displayName ?? '');
const previewResult = ref<DeckPreview | null>(null);
const stage = ref<'import' | 'commander'>(selectedCommanders.value.length ? 'commander' : 'import');
const creating = ref(false);
const pageError = ref<ActivationErrorEntry | null>(null);

onMounted(() => track('quick_start_viewed'));

const commanderCandidates = computed<CommanderPick[]>(() => {
  const candidates = previewResult.value?.CommanderCandidates ?? [];
  return candidates.length ? candidates : selectedCommanders.value;
});

const legalCommanderSelection = computed(() => {
  if (selectedCommanders.value.length === 1) return true;
  if (selectedCommanders.value.length !== 2) return false;
  // CommanderReview is the only mutation path and has already validated the pair.
  return true;
});

const canStartTable = computed(() => Boolean(
  previewResult.value?.CanContinue
  && previewResult.value.BlockingErrors.length === 0
  && legalCommanderSelection.value
  && !creating.value,
));

function draftSnapshot(): QuickStartDraft {
  return {
    deckText: deckText.value,
    sourceURL: sourceURL.value,
    corrections: { ...corrections.value },
    selectedCommanders: [...selectedCommanders.value],
    displayName: displayName.value,
  };
}

function persistDraft(): void {
  saveDraft(draftSnapshot());
}

watch(displayName, persistDraft);

function handleDeckChange(change: DeckImportChange): void {
  deckText.value = change.text;
  sourceURL.value = change.sourceURL;
  corrections.value = { ...change.corrections };
  persistDraft();
}

function handleDeckImportStarted(): void {
  track('deck_import_started');
}

function handlePreviewResolved(preview: DeckPreview): void {
  previewResult.value = preview;
  pageError.value = null;
}

function handlePanelError(_error: ActivationErrorEntry): void {
  // DeckImportPanel owns the one preview error region. Keeping this listener
  // explicit prevents a second page-level copy block from leaking into view.
  pageError.value = null;
}

function handleCommanderChange(selected: CommanderPick[]): void {
  selectedCommanders.value = [...selected];
  persistDraft();
}

function correctedDeckText(): string {
  if (sourceURL.value && previewResult.value) {
    return previewResult.value.Entries.map((entry) => `${entry.Quantity} ${entry.Name}`).join('\n');
  }
  const lines = deckText.value.split(/\r?\n/);
  for (const [sourceLine, correctedName] of Object.entries(corrections.value)) {
    const index = Number(sourceLine) - 1;
    if (index < 0 || index >= lines.length) continue;
    if (!correctedName) {
      lines[index] = '';
      continue;
    }
    const quantityPrefix = lines[index].match(/^\s*\d+\s*(?:x|×)?\s*[,;:\-]?\s*/i)?.[0] ?? '';
    lines[index] = `${quantityPrefix}${correctedName}`;
  }
  return lines.filter((line) => line.length > 0).join('\n');
}

async function handleStartTable(): Promise<void> {
  if (!canStartTable.value) return;
  creating.value = true;
  pageError.value = null;
  let action: 'guest' | 'create' = auth.isAuthenticated ? 'create' : 'guest';
  try {
    if (!auth.isAuthenticated) {
      await auth.createGuestSession({
        displayName: displayName.value || undefined,
        sessionID,
      });
    }

    const profile = auth.profile;
    if (!profile) throw new Error('guest session did not establish a profile');
    action = 'create';
    const gameID = crypto.randomUUID();
    const payload = {
      ID: gameID,
      Handle: 'Commander table',
      FormatID: 'EDH',
      SessionID: sessionID,
      Turn: { Player: profile.Username, Phase: 'MAIN', Number: 1, Priority: profile.Username },
      Players: [{
        UserID: profile.ID,
        User: profile.Username,
        GameID: gameID,
        Life: 40,
        Decklist: correctedDeckText(),
        Commander: selectedCommanders.value.map(({ ID, Name }) => ({ ID, Name })),
        Library: [],
        Graveyard: [],
        Exiled: [],
        Battlefield: [],
        Hand: [],
        Revealed: [],
        Controlled: [],
        Counters: [],
      }],
    } as const;

    track('game_create_started');
    const createdID = await games.createGame(payload);
    if (!createdID) throw new Error('createGame returned no game');
    clearDraft();
    await router.push({ name: 'board', params: { id: createdID } });
  } catch (error: unknown) {
    pageError.value = resolveActivationError(errorWithFallbackCode(error, action === 'guest' ? 'guest_session_error' : 'create_error'));
    persistDraft();
  } finally {
    creating.value = false;
  }
}

function errorWithFallbackCode(error: unknown, code: 'guest_session_error' | 'create_error'): unknown {
  if (typeof error === 'object' && error !== null && 'graphQLErrors' in error) return error;
  return { graphQLErrors: [{ extensions: { code } }] };
}
</script>

<style scoped lang="scss">
@use '@/styles/breakpoints' as bp;

.quick-start {
  box-sizing: border-box;
  width: 100%;
  min-width: 0;
  margin: 0 auto;
  padding: 48px 16px;
  display: grid;
  gap: 24px;
  overflow-x: hidden;
}

.display-heading {
  margin: 0;
  color: var(--vedh-text);
  font-size: 28px;
  font-weight: 600;
  line-height: 1.2;
  opacity: 0.9;
}

.quick-start-layout,
.finish-pane,
.display-name-field,
.actions,
.page-error {
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 16px;
}

.display-name-field input {
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

button {
  border-radius: 10px;
  padding: 12px 16px;
  font: inherit;
  font-weight: 600;
}

button.primary {
  border: 0;
  background: var(--vedh-primary-gradient);
  color: var(--vedh-primary-contrast);
}

button.secondary {
  border: 1px solid var(--vedh-border);
  background: var(--vedh-surface);
  color: var(--vedh-text);
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-error {
  padding: 16px;
  border: 1px solid var(--vedh-danger);
  border-radius: 12px;
  color: var(--vedh-danger);
}

.page-error h2,
.page-error p {
  margin: 0;
}

@media (min-width: bp.$breakpoint-tablet) {
  .quick-start {
    padding-right: 32px;
    padding-left: 32px;
  }

  .quick-start-layout {
    grid-template-columns: minmax(0, 2fr) minmax(280px, 1fr);
    gap: 32px;
  }

  .quick-start-deck-import--wide {
    grid-column: 1 / -1;
  }
}
</style>
