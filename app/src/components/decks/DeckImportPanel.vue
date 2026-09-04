<template>
  <section
    ref="panelElement"
    class="deck-import-panel"
    data-testid="deck-import-panel"
    :data-persistence-key="persistenceKey"
    :aria-label="`${context.gameLabel} ${context.formatLabel} deck import`"
  >
    <div class="input-pane">
      <header class="panel-heading">
        <h2>Paste your decklist</h2>
        <p>Paste a deck export or plain-text list, or use a supported public deck URL. We'll check every card and get you to commander selection.</p>
        <dl class="deck-context" data-testid="deck-import-context">
          <div>
            <dt>Game</dt>
            <dd>{{ context.gameLabel }}</dd>
          </div>
          <div>
            <dt>Format</dt>
            <dd>{{ context.formatLabel }}</dd>
          </div>
        </dl>
      </header>

      <div class="source-tabs" role="tablist" aria-label="Deck source">
        <button
          id="deck-source-paste-tab"
          ref="pasteTabElement"
          type="button"
          class="secondary source-tab"
          :class="{ 'is-selected': activeSource === 'paste' }"
          :aria-selected="activeSource === 'paste'"
          aria-controls="deck-source-paste-panel"
          :tabindex="activeSource === 'paste' ? 0 : -1"
          role="tab"
          data-testid="paste-tab"
          @click="activateSource('paste')"
          @keydown="handleSourceTabKeydown($event, 'paste')"
        >
          <span class="source-tab-marker" aria-hidden="true">{{ activeSource === 'paste' ? '✓' : '' }}</span>
          <span>Paste</span>
        </button>
        <button
          id="deck-source-url-tab"
          ref="urlTabElement"
          type="button"
          class="secondary source-tab"
          :class="{ 'is-selected': activeSource === 'url' }"
          :aria-selected="activeSource === 'url'"
          aria-controls="deck-source-url-panel"
          :tabindex="activeSource === 'url' ? 0 : -1"
          role="tab"
          data-testid="source-url-tab"
          @click="activateSource('url')"
          @keydown="handleSourceTabKeydown($event, 'url')"
        >
          <span class="source-tab-marker" aria-hidden="true">{{ activeSource === 'url' ? '✓' : '' }}</span>
          <span>Public URL</span>
        </button>
      </div>

      <div class="panel-form">
        <div
          id="deck-source-paste-panel"
          class="source-panel"
          role="tabpanel"
          aria-labelledby="deck-source-paste-tab"
          data-testid="paste-panel"
          :hidden="activeSource !== 'paste'"
        >
          <label class="stacked">
            <span>Decklist</span>
            <textarea
              ref="deckTextElement"
              v-model="deckText"
              class="deck-textarea"
              data-testid="deck-text"
              aria-describedby="deck-text-format-guidance"
              rows="10"
              wrap="soft"
              spellcheck="false"
            ></textarea>
            <p id="deck-text-format-guidance" class="input-guidance" data-testid="paste-format-guidance">
              Paste Moxfield or Archidekt text exports, CSV, or plain text. Accepted lines include
              <code>1 Sol Ring</code>, <code>1x Sol Ring</code>, <code>1,Sol Ring</code>,
              <code>1, Sol Ring</code>, quoted CSV names, or a card name without a quantity.
            </p>
          </label>
        </div>
        <div
          id="deck-source-url-panel"
          class="source-panel"
          role="tabpanel"
          aria-labelledby="deck-source-url-tab"
          data-testid="source-url-panel"
          :hidden="activeSource !== 'url'"
        >
          <label class="stacked">
            <span>Public deck URL</span>
            <input
              v-model="sourceURL"
              data-testid="source-url"
              type="url"
              aria-describedby="deck-url-site-guidance"
              placeholder="https://archidekt.com/decks/…"
              @keydown.enter.prevent="submitPreview"
            />
            <p id="deck-url-site-guidance" class="input-guidance" data-testid="url-site-guidance">
              Public Archidekt deck URLs are supported. URL import may be disabled by the server operator.
            </p>
          </label>
        </div>

        <button
          class="primary"
          type="button"
          data-testid="deck-preview-submit"
          :disabled="!canSubmit"
          @click="submitPreview"
        >
          {{ activeSource === 'url' ? 'Import from URL' : 'Preview deck' }}
        </button>
      </div>

      <div v-if="loading" class="loading-state" aria-live="polite">
        <span class="spinner" data-testid="preview-spinner" aria-hidden="true"></span>
        <span>Checking your deck</span>
      </div>

      <section v-if="activationError" class="activation-error" data-testid="activation-error" role="alert">
        <h3>{{ activationError.title }}</h3>
        <p>{{ activationError.body }}</p>
        <button class="secondary" type="button" @click="handleErrorAffordance">
          {{ activationError.affordance }}
        </button>
      </section>
      <div v-if="Object.keys(corrections).length && !previewResult" class="restored-corrections">
        <p v-for="(name, line) in corrections" :key="line">
          Line {{ line }}: Using {{ name || 'removed line' }}
        </p>
      </div>
    </div>

    <div class="preview-pane">
      <section
        v-if="previewResult"
        class="preview-totals"
        data-testid="preview-totals"
        :aria-live="pasteReady ? 'off' : 'polite'"
      >
        <p class="ready-line">{{ readyLine }}</p>
        <p>{{ warningLine }}</p>
        <p :class="{ blocking: previewResult.BlockingErrors.length > 0 }">{{ blockingLine }}</p>
        <ul v-if="previewResult.Warnings.length || previewResult.BlockingErrors.length" class="message-list">
          <li
            v-for="message in [...previewResult.Warnings, ...previewResult.BlockingErrors]"
            :key="message"
            class="truncated"
            :title="message"
          >
            {{ truncateDisplay(message) }}
          </li>
        </ul>
      </section>

      <section
        v-if="pasteSuccessAcknowledgement"
        :key="pasteSuccessAcknowledgement.revision"
        class="paste-import-success"
        data-testid="paste-import-success"
        role="status"
        aria-live="polite"
        aria-atomic="true"
      >
        <span class="paste-import-success-icon" aria-hidden="true">✓</span>
        <div class="paste-import-success-copy">
          <strong>Deck import successful</strong>
          <p>{{ pasteSuccessAcknowledgement.detail }}</p>
        </div>
      </section>

      <section v-if="previewResult?.Unresolved.length" class="unresolved-section">
        <h3>{{ unresolvedHeading }}</h3>
        <p class="unresolved-guidance" data-testid="unresolved-guidance">
          Choose a replacement to include each unresolved card. You can continue without one,
          but that line will be left out of your deck.
        </p>
        <ul class="typeahead unresolved-list" data-testid="unresolved-list">
          <li
            v-for="issue in previewResult.Unresolved"
            :key="issue.SourceLine"
            class="unresolved-row"
            :class="{ 'is-reviewed': hasExplicitCorrection(issue.SourceLine) }"
            data-testid="unresolved-row"
          >
            <div class="row-heading">
              <div class="original-card">
                <span class="row-label">Original</span>
                <strong class="truncated" :title="issue.RawLine">{{ truncateDisplay(issue.RawLine) }}</strong>
              </div>
              <button
                type="button"
                class="secondary row-remove"
                data-testid="remove-unresolved-line"
                :aria-label="`Leave ${issue.Name} out of the deck`"
                @click="removeIssue(issue)"
              >
                Leave line out
              </button>
            </div>

            <div
              v-if="hasExplicitCorrection(issue.SourceLine)"
              class="correction-confirmation"
              data-testid="correction-confirmation"
              role="status"
              aria-live="polite"
              aria-atomic="true"
            >
              <span class="row-label">{{ corrections[issue.SourceLine] ? 'Replacement selected' : 'Line reviewed' }}</span>
              <p
                class="correction-route"
                :aria-label="correctionAnnouncement(issue)"
              >
                <span class="truncated" :title="issue.Name" aria-hidden="true">{{ truncateDisplay(issue.Name) }}</span>
                <span aria-hidden="true">→</span>
                <strong class="truncated" :title="corrections[issue.SourceLine] || 'Left out of deck'" aria-hidden="true">
                  {{ corrections[issue.SourceLine] || 'Left out of deck' }}
                </strong>
              </p>
              <div class="correction-actions">
                <button
                  v-if="corrections[issue.SourceLine]"
                  type="button"
                  class="secondary compact-control"
                  :aria-label="`Change replacement for ${issue.Name}`"
                  @click="focusReplacementOptions(issue.SourceLine)"
                >
                  Change
                </button>
                <button
                  type="button"
                  class="secondary compact-control"
                  :aria-label="`Undo review choice for ${issue.Name}`"
                  data-testid="undo-correction"
                  @click="undoCorrection(issue.SourceLine)"
                >
                  Undo
                </button>
              </div>
            </div>

            <fieldset class="replacement-group">
              <legend>Suggested replacements</legend>
              <div class="replacement-options">
                <label
                v-for="candidate in issue.Candidates.slice(0, 3)"
                :key="candidate.Name"
                  class="replacement-option"
                  :class="{ 'is-selected': corrections[issue.SourceLine] === candidate.Name }"
                  :title="candidate.Name"
                  data-testid="suggestion-option"
                >
                  <input
                    type="radio"
                    :name="`replacement-${issue.SourceLine}`"
                    :value="candidate.Name"
                    :checked="corrections[issue.SourceLine] === candidate.Name"
                    :aria-label="`Replace ${issue.Name} with ${candidate.Name}${candidate.LowConfidence ? ', closest available match' : ''}`"
                    data-testid="suggestion-control"
                    @change="applyCorrection(issue.SourceLine, candidate.Name)"
                  />
                  <span class="replacement-option-copy">
                    <span class="truncated">Replace with <strong>{{ truncateDisplay(candidate.Name) }}</strong></span>
                    <small v-if="candidate.LowConfidence">Closest available match</small>
                  </span>
                </label>
              </div>
            </fieldset>

            <label class="stacked manual-search">
              <span>Search for a different card</span>
              <input
                data-testid="manual-card-name"
                type="text"
                autocomplete="off"
                :value="manualQueries[issue.SourceLine] ?? ''"
                @input="handleManualInput(issue.SourceLine, $event)"
                @keydown.enter.prevent="applyManualCorrection(issue.SourceLine)"
              />
              <ul
                v-if="manualSearching.has(issue.SourceLine) || manualResults[issue.SourceLine]?.length"
                class="typeahead manual-results"
                role="listbox"
              >
                <li v-if="manualSearching.has(issue.SourceLine)" class="search-hint" role="option" aria-disabled="true">
                  Searching…
                </li>
                <li
                  v-for="card in manualResults[issue.SourceLine] ?? []"
                  v-else
                  :key="card.ID"
                  role="option"
                  @mousedown.prevent="applyCorrection(issue.SourceLine, card.Name)"
                >
                  <button type="button" class="manual-result truncated" :title="card.Name">
                    {{ truncateDisplay(card.Name) }}
                  </button>
                </li>
              </ul>
            </label>
          </li>
        </ul>
      </section>

      <button
        v-if="previewResult && !continued"
        class="primary continue-button"
        type="button"
        data-testid="continue-to-commanders"
        :disabled="!canContinue"
        @click="continueToCommanders"
      >
        {{ continueLabel }}
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue';
import { PREVIEW_DECK_MUTATION } from '../../graphql/mutations';
import { SEARCH_CARDS_QUERY } from '../../graphql/queries';
import { apolloClient } from '../../services/apollo';
import {
  resolveActivationError,
  type ActivationErrorEntry,
} from '../../services/activationErrors';
import { QUICK_START_DRAFT_KEY, readDraft } from '../../services/quickStartDraft';
import type { DeckImportContext } from './deckImportContext';
import type {
  DeckImportIssue,
  DeckPreview,
  Card,
  PreviewDeckMutation,
  PreviewDeckMutationVariables,
} from '../../types/generated';

export type DeckImportChange = {
  text: string;
  sourceURL: string;
  corrections: Record<number, string>;
  decklist: string;
};

const props = withDefaults(defineProps<{
  initialText?: string;
  initialSourceURL?: string;
  sessionId: string;
  persistenceKey?: string;
  context: DeckImportContext;
}>(), {
  initialText: '',
  initialSourceURL: '',
  persistenceKey: undefined,
});

const emit = defineEmits<{
  (event: 'submit'): void;
  (event: 'preview-resolved', preview: DeckPreview): void;
  (event: 'change', value: DeckImportChange): void;
  (event: 'error', value: ActivationErrorEntry): void;
  (event: 'continue', preview: DeckPreview): void;
}>();

const deckText = ref(props.initialText);
const sourceURL = ref(props.initialSourceURL);
const activeSource = ref<'paste' | 'url'>(props.initialSourceURL && !props.initialText ? 'url' : 'paste');
const restoredCorrections = props.persistenceKey === QUICK_START_DRAFT_KEY
  ? readDraft()?.corrections ?? {}
  : {};
const corrections = reactive<Record<number, string>>({ ...restoredCorrections });
const manualQueries = reactive<Record<number, string>>({});
const manualResults = reactive<Record<number, Card[]>>({});
const manualSearching = reactive(new Set<number>());
const manualDebounces = new Map<number, number>();
const loading = ref(false);
const previewResult = ref<DeckPreview | null>(null);
const previewInputSource = ref<'paste' | 'url' | null>(null);
const activationError = ref<ActivationErrorEntry | null>(null);
const continued = ref(false);
const deckTextElement = ref<HTMLTextAreaElement | null>(null);
const panelElement = ref<HTMLElement | null>(null);
const pasteTabElement = ref<HTMLButtonElement | null>(null);
const urlTabElement = ref<HTMLButtonElement | null>(null);
let requestSequence = 0;
let previewRevision = 0;

type PasteSuccessAcknowledgement = {
  revision: number;
  detail: string;
};

const pasteSuccessAcknowledgement = ref<PasteSuccessAcknowledgement | null>(null);

const hasCurrentInput = computed(() => (
  activeSource.value === 'url' ? sourceURL.value.trim().length > 0 : deckText.value.length > 0
));
const canSubmit = computed(() => hasCurrentInput.value && !loading.value);
const canContinue = computed(() => Boolean(
  previewResult.value?.CanContinue && previewResult.value.BlockingErrors.length === 0,
));

function hasExplicitCorrection(sourceLine: number): boolean {
  return sourceLine in corrections;
}

const pendingUnresolvedCount = computed(() => {
  const result = previewResult.value;
  if (!result) return 0;
  return result.Unresolved.filter((issue) => !hasExplicitCorrection(issue.SourceLine)).length;
});
function unresolvedQuantity(issue: DeckImportIssue): number {
  return previewResult.value?.Entries.find((entry) => entry.SourceLine === issue.SourceLine)?.Quantity ?? 1;
}
const pendingUnresolvedQuantity = computed(() => {
  const result = previewResult.value;
  if (!result) return 0;
  return result.Unresolved.reduce((count, issue) => (
    count + (hasExplicitCorrection(issue.SourceLine) ? 0 : unresolvedQuantity(issue))
  ), 0);
});
const correctedUnresolvedCount = computed(() => {
  const result = previewResult.value;
  if (!result) return 0;
  return result.Unresolved.filter((issue) => (
    hasExplicitCorrection(issue.SourceLine) && corrections[issue.SourceLine].length > 0
  )).length;
});
const correctedUnresolvedQuantity = computed(() => {
  const result = previewResult.value;
  if (!result) return 0;
  return result.Unresolved.reduce((count, issue) => (
    count + (
      hasExplicitCorrection(issue.SourceLine) && corrections[issue.SourceLine].length > 0
        ? unresolvedQuantity(issue)
        : 0
    )
  ), 0);
});
const removedUnresolvedCount = computed(() => {
  const result = previewResult.value;
  if (!result) return 0;
  return result.Unresolved.filter((issue) => (
    hasExplicitCorrection(issue.SourceLine) && corrections[issue.SourceLine].length === 0
  )).length;
});
const pasteReady = computed(() => Boolean(
  !loading.value
  && !activationError.value
  && previewInputSource.value === 'paste'
  && canContinue.value
  && pendingUnresolvedCount.value === 0,
));

const readyCount = computed(() => {
  const result = previewResult.value;
  if (!result) return 0;
  return Math.max(result.CardCount + correctedUnresolvedQuantity.value, 0);
});
const readyLine = computed(() => {
  const count = Math.max(readyCount.value + pendingUnresolvedQuantity.value, 0);
  const needs = pendingUnresolvedQuantity.value;
  const base = `${readyCount.value} of ${count} ${count === 1 ? 'card' : 'cards'} ready`;
  return needs === 0 ? base : `${base} — ${needs} ${needs === 1 ? 'card needs' : 'cards need'} a quick look`;
});
const warningLine = computed(() => {
  const count = previewResult.value?.Warnings.length ?? 0;
  return count === 0 ? 'No warnings' : `${count} ${count === 1 ? 'warning' : 'warnings'}`;
});
const blockingLine = computed(() => {
  const count = previewResult.value?.BlockingErrors.length ?? 0;
  return count === 0 ? 'No blocking errors' : `${count} blocking ${count === 1 ? 'error' : 'errors'}`;
});
const unresolvedHeading = computed(() => {
  const count = pendingUnresolvedCount.value;
  if (count === 0) {
    return 'All unresolved lines reviewed';
  }
  return `Choose ${count} ${count === 1 ? 'replacement' : 'replacements'}`;
});
const continueLabel = computed(() => {
  const count = pendingUnresolvedQuantity.value;
  if (count === 0) return 'Continue to commanders';
  return `Continue without ${count} ${count === 1 ? 'card' : 'cards'}`;
});
const pasteSuccessDetail = computed(() => {
  const cardCount = readyCount.value;
  const correctionCount = correctedUnresolvedCount.value;
  const removedCount = removedUnresolvedCount.value;
  const contextParts = [`${cardCount} ${cardCount === 1 ? 'card is' : 'cards are'} ready`];
  if (correctionCount > 0) {
    contextParts.push(`${correctionCount} ${correctionCount === 1 ? 'correction' : 'corrections'} applied`);
  }
  if (removedCount > 0) {
    contextParts.push(`${removedCount} ${removedCount === 1 ? 'line' : 'lines'} removed`);
  }
  return `${contextParts.join(' · ')}. Continue to choose your commander.`;
});

watch(pasteReady, (ready) => {
  if (!ready) {
    pasteSuccessAcknowledgement.value = null;
    return;
  }
  pasteSuccessAcknowledgement.value = {
    revision: previewRevision,
    detail: pasteSuccessDetail.value,
  };
}, { flush: 'post' });

function currentChange(): DeckImportChange {
  return {
    text: deckText.value,
    sourceURL: sourceURL.value,
    corrections: { ...corrections },
    decklist: preparedDecklist(),
  };
}

function preparedDecklist(): string {
  if (activeSource.value === 'url' && previewResult.value) {
    return previewResult.value.Entries.flatMap((entry) => {
      if (!(entry.SourceLine in corrections)) return [`${entry.Quantity} ${entry.Name}`];
      const correctedName = corrections[entry.SourceLine];
      return correctedName ? [`${entry.Quantity} ${correctedName}`] : [];
    }).join('\n');
  }
  const lines = deckText.value.split(/\r?\n/);
  for (const [sourceLine, correctedName] of Object.entries(corrections)) {
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

watch([deckText, sourceURL, corrections], () => emit('change', currentChange()), {
  deep: true,
  immediate: true,
});

async function submitPreview(): Promise<void> {
  if (!hasCurrentInput.value) return;
  emit('submit');
  const sequence = ++requestSequence;
  loading.value = true;
  activationError.value = null;
  previewResult.value = null;
  previewInputSource.value = null;
  continued.value = false;

  const submittedSource = activeSource.value;
  const input = submittedSource === 'url'
    ? { sourceURL: sourceURL.value, sessionID: props.sessionId }
    : { text: deckText.value, sessionID: props.sessionId };

  try {
    const { data } = await apolloClient.mutate<PreviewDeckMutation, PreviewDeckMutationVariables>({
      mutation: PREVIEW_DECK_MUTATION,
      variables: { input },
    });
    if (sequence !== requestSequence) return;
    const result = data?.previewDeck ?? null;
    if (!result) throw new Error('previewDeck returned no result');
    previewRevision += 1;
    previewInputSource.value = submittedSource;
    previewResult.value = result;
    emit('preview-resolved', result);
    emit('change', currentChange());
  } catch (error: unknown) {
    if (sequence !== requestSequence) return;
    const entry = resolveActivationError(error);
    activationError.value = entry;
    previewResult.value = null;
    emit('error', entry);
  } finally {
    if (sequence === requestSequence) loading.value = false;
  }
}

function applyCorrection(sourceLine: number, name: string): void {
  corrections[sourceLine] = name;
}

function clearCorrection(sourceLine: number): void {
  delete corrections[sourceLine];
}

function correctionAnnouncement(issue: DeckImportIssue): string {
  const correction = corrections[issue.SourceLine];
  return correction
    ? `${issue.Name} is replaced by ${correction}`
    : `${issue.Name} will be left out of the deck`;
}

function focusReplacementOptions(sourceLine: number): void {
  const selected = panelElement.value?.querySelector<HTMLInputElement>(
    `input[name="replacement-${sourceLine}"]:checked`,
  );
  const first = panelElement.value?.querySelector<HTMLInputElement>(
    `input[name="replacement-${sourceLine}"]`,
  );
  (selected ?? first)?.focus();
}

async function undoCorrection(sourceLine: number): Promise<void> {
  clearCorrection(sourceLine);
  await nextTick();
  focusReplacementOptions(sourceLine);
}

function removeIssue(issue: DeckImportIssue): void {
  corrections[issue.SourceLine] = '';
}

function handleManualInput(sourceLine: number, event: Event): void {
  const value = (event.target as HTMLInputElement).value;
  manualQueries[sourceLine] = value;
  const pending = manualDebounces.get(sourceLine);
  if (pending !== undefined) window.clearTimeout(pending);
  if (value.trim().length < 2) {
    manualResults[sourceLine] = [];
    manualSearching.delete(sourceLine);
    return;
  }
  manualDebounces.set(sourceLine, window.setTimeout(() => {
    void runManualSearch(sourceLine, value.trim());
  }, 150));
}

async function runManualSearch(sourceLine: number, query: string): Promise<void> {
  manualSearching.add(sourceLine);
  try {
    const { data } = await apolloClient.query<{ search?: Card[] }>({
      query: SEARCH_CARDS_QUERY,
      variables: { name: `%${query}%` },
      fetchPolicy: 'no-cache',
    });
    manualResults[sourceLine] = (data?.search ?? []).slice(0, 8);
  } catch {
    manualResults[sourceLine] = [];
  } finally {
    manualSearching.delete(sourceLine);
  }
}

function applyManualCorrection(sourceLine: number): void {
  const value = manualQueries[sourceLine]?.trim();
  if (value) applyCorrection(sourceLine, value);
}

function continueToCommanders(): void {
  if (!previewResult.value || !canContinue.value) return;
  continued.value = true;
  emit('continue', previewResult.value);
}

type DeckSource = 'paste' | 'url';

function activateSource(source: DeckSource, moveFocus = false): void {
  activeSource.value = source;
  if (!moveFocus) return;
  void nextTick(() => {
    (source === 'paste' ? pasteTabElement.value : urlTabElement.value)?.focus();
  });
}

function handleSourceTabKeydown(event: KeyboardEvent, source: DeckSource): void {
  let nextSource: DeckSource | null = null;
  if (event.key === 'ArrowLeft' || event.key === 'ArrowUp') {
    nextSource = source === 'paste' ? 'url' : 'paste';
  } else if (event.key === 'ArrowRight' || event.key === 'ArrowDown') {
    nextSource = source === 'paste' ? 'url' : 'paste';
  } else if (event.key === 'Home') {
    nextSource = 'paste';
  } else if (event.key === 'End') {
    nextSource = 'url';
  }
  if (!nextSource) return;
  event.preventDefault();
  activateSource(nextSource, true);
}

async function handleErrorAffordance(): Promise<void> {
  if (activationError.value?.affordance === 'Paste instead') activateSource('paste');
  if (activationError.value?.affordance === 'Try again') {
    await submitPreview();
    return;
  }
  await nextTick();
  deckTextElement.value?.focus();
}

function truncateDisplay(value: string, maximum = 48): string {
  const Segmenter = typeof Intl !== 'undefined'
    ? (Intl as typeof Intl & { Segmenter?: new (...args: unknown[]) => { segment: (text: string) => Iterable<{ segment: string }> } }).Segmenter
    : undefined;
  const segments = Segmenter
    ? Array.from(new Segmenter(undefined, { granularity: 'grapheme' } as never).segment(value), (part) => part.segment)
    : Array.from(value);
  return segments.length <= maximum ? value : `${segments.slice(0, maximum).join('')}…`;
}

onBeforeUnmount(() => {
  for (const timer of manualDebounces.values()) window.clearTimeout(timer);
});

defineExpose({ submitPreview, currentChange });
</script>

<style scoped lang="scss">
@use '@/styles/breakpoints' as bp;

.deck-import-panel {
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 24px;
  padding: 24px;
  border: 1px solid var(--vedh-border);
  border-radius: 20px;
  background: var(--vedh-panel);
  overflow-x: hidden;
}

.input-pane,
.preview-pane,
.panel-form,
.stacked,
.panel-heading,
.unresolved-section {
  min-width: 0;
  display: grid;
  gap: 8px;
}

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

p {
  font-size: 16px;
  font-weight: 400;
  line-height: 1.5;
}

.input-guidance {
  color: var(--vedh-muted);
}

.input-guidance code {
  color: var(--vedh-text);
  font: inherit;
  white-space: nowrap;
}

.deck-context,
.deck-context div {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 8px;
}

.deck-context {
  margin: 8px 0 0;
  column-gap: 24px;
}

.deck-context dt {
  color: var(--vedh-muted);
}

.deck-context dd {
  margin: 0;
  font-weight: 600;
}

.row-heading,
.loading-state,
.replacement-options,
.correction-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.source-tabs {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  padding: 4px;
  border: 1px solid var(--vedh-border);
  border-radius: 12px;
  background: rgba(255, 244, 237, 0.05);
}

.source-tab {
  min-width: 0;
  min-height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border: 2px solid transparent;
  background: transparent;
}

.source-tab.is-selected {
  border-color: var(--vedh-border-strong);
  background: var(--vedh-surface-strong);
  box-shadow: inset 0 -3px 0 var(--vedh-text);
}

.source-tab-marker {
  flex: 0 0 16px;
  width: 16px;
  font-weight: 600;
  text-align: center;
}

.source-tab:focus-visible {
  outline: 3px solid var(--vedh-primary);
  outline-offset: 2px;
}

.source-panel {
  min-width: 0;
  display: grid;
  gap: 8px;
}

.source-panel[hidden] {
  display: none;
}

.row-heading {
  justify-content: space-between;
  flex-wrap: nowrap;
}

.original-card {
  min-width: 0;
  display: grid;
  gap: 4px;
}

.row-label,
.unresolved-guidance {
  color: var(--vedh-muted);
}

.row-label {
  font-size: 14px;
  line-height: 1.4;
}

textarea,
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

.deck-textarea {
  height: 224px;
  max-height: 224px;
  resize: none;
  overflow-y: auto;
  overflow-x: hidden;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  word-break: break-word;
}

button {
  border-radius: 10px;
  padding: 12px 16px;
  font: inherit;
  font-weight: 600;
  cursor: pointer;
}

button.primary {
  border: 0;
  background: var(--vedh-primary-gradient);
  color: var(--vedh-primary-contrast);
}

button.secondary,
.replacement-option {
  border: 1px solid var(--vedh-border);
  background: rgba(255, 244, 237, 0.05);
  color: var(--vedh-text);
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid var(--vedh-border-strong);
  border-top-color: var(--vedh-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.activation-error,
.preview-totals,
.paste-import-success {
  display: grid;
  gap: 8px;
  padding: 16px;
  border-radius: 12px;
  background: var(--vedh-surface);
  border: 1px solid var(--vedh-border);
}

.paste-import-success {
  grid-template-columns: 40px minmax(0, 1fr);
  align-items: center;
  border-color: var(--vedh-border-strong);
  background: var(--vedh-surface-strong);
  animation: paste-import-success-arrive 480ms cubic-bezier(0.2, 0.8, 0.2, 1) both;
}

.paste-import-success-icon {
  width: 40px;
  height: 40px;
  display: grid;
  place-items: center;
  border: 2px solid var(--vedh-border-strong);
  border-radius: 50%;
  background: var(--vedh-panel-strong);
  font-size: 24px;
  font-weight: 600;
  line-height: 1;
  animation: paste-import-success-check 560ms cubic-bezier(0.2, 0.9, 0.2, 1.2) both;
}

.paste-import-success-copy {
  min-width: 0;
  display: grid;
  gap: 4px;
}

.paste-import-success-copy strong {
  font-size: 16px;
  font-weight: 600;
  line-height: 1.4;
}

.paste-import-success-copy p {
  color: var(--vedh-muted);
}

.activation-error,
.blocking {
  color: var(--vedh-danger);
}

.typeahead {
  max-height: 200px;
  overflow: auto;
}

.unresolved-list,
.message-list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.unresolved-row {
  min-width: 0;
  display: grid;
  gap: 8px;
  padding: 16px;
  border-bottom: 1px solid var(--vedh-border);
}

.unresolved-row.is-reviewed {
  background: rgba(255, 244, 237, 0.03);
}

.unresolved-row:last-child {
  border-bottom: 0;
}

.replacement-group {
  min-width: 0;
  margin: 0;
  padding: 0;
  border: 0;
}

.replacement-group legend {
  margin-bottom: 4px;
  padding: 0;
  font-weight: 600;
}

.replacement-options {
  align-items: stretch;
}

.replacement-option {
  min-width: 0;
  max-width: 100%;
  min-height: 44px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 10px;
  color: var(--vedh-primary);
  cursor: pointer;
}

.replacement-option.is-selected {
  border-color: var(--vedh-primary);
  background: var(--vedh-surface-strong);
}

.replacement-option input {
  flex: 0 0 auto;
  width: 18px;
  min-width: 18px;
  height: 18px;
  margin: 0;
  padding: 0;
  accent-color: var(--vedh-primary);
}

.replacement-option input:focus-visible {
  outline: 3px solid var(--vedh-primary);
  outline-offset: 4px;
}

.replacement-option-copy {
  min-width: 0;
  display: grid;
  gap: 2px;
}

.replacement-option-copy small {
  color: var(--vedh-muted);
  font-size: 12px;
  line-height: 1.4;
}

.row-remove,
.compact-control {
  min-height: 44px;
  padding: 8px 12px;
}

.correction-actions {
  flex-wrap: nowrap;
}

.compact-control {
  flex: 1 1 0;
  min-width: 0;
}

.manual-results {
  margin: 0;
  padding: 0;
  list-style: none;
}

.manual-result {
  box-sizing: border-box;
  width: 100%;
  border: 0;
  background: transparent;
  color: var(--vedh-text);
  text-align: left;
}

.truncated {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.icon-control {
  flex: 0 0 44px;
  min-width: 44px;
  min-height: 44px;
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--vedh-text);
}

.correction-confirmation {
  min-width: 0;
  display: grid;
  gap: 8px;
  padding: 12px;
  border: 1px solid var(--vedh-border-strong);
  border-radius: 10px;
  background: var(--vedh-surface-strong);
}

.correction-route {
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: 8px;
}

.correction-route strong {
  color: var(--vedh-text);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@keyframes paste-import-success-arrive {
  from {
    opacity: 0;
    transform: translateY(8px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes paste-import-success-check {
  from { transform: scale(0.72) rotate(-8deg); }
  65% { transform: scale(1.12) rotate(2deg); }
  to { transform: scale(1) rotate(0); }
}

@media (prefers-reduced-motion: reduce) {
  .paste-import-success,
  .paste-import-success-icon {
    animation: none;
  }
}

@media (min-width: bp.$breakpoint-tablet) {
  .deck-import-panel {
    grid-template-columns: minmax(0, 3fr) minmax(0, 2fr);
    gap: 32px;
  }
}
</style>
