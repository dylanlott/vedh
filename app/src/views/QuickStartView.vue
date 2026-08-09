<template>
  <section class="quick-start">
    <h1 class="display-heading">Paste a deck. Start a table.</h1>

    <form class="paste-panel" @submit.prevent="handlePreview">
      <label class="stacked">
        <span>Paste your decklist</span>
        <textarea
          v-model="deckText"
          rows="10"
          placeholder="Paste a list from any format you already use, or drop in a public deck URL below. We'll check every card and get you to commander selection."
        ></textarea>
      </label>

      <label class="stacked">
        <span>Display name (optional)</span>
        <input
          v-model="displayName"
          type="text"
          maxlength="64"
          placeholder="Optional — skip it and we'll give you a name like 'Brave Sliver'."
        />
      </label>

      <button class="primary" type="submit" data-testid="preview-submit" :disabled="!canSubmitPreview">
        {{ previewing ? 'Checking your deck…' : 'Preview deck' }}
      </button>

      <p v-if="previewError" class="error-text">{{ previewError }}</p>
    </form>

    <section v-if="previewResult" class="preview-totals">
      <p>{{ totalsLine }}</p>
    </section>

    <footer v-if="previewResult" class="actions">
      <button class="primary" type="button" data-testid="start-table" :disabled="!canStartTable" @click="handleStartTable">
        {{ creating ? 'Starting your table…' : 'Start my table' }}
      </button>
      <p v-if="createError" class="error-text">{{ createError }}</p>
    </footer>
  </section>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { apolloClient } from '../services/apollo';
import { PREVIEW_DECK_MUTATION } from '../graphql/mutations';
import { getSessionID } from '../services/productEvents';
import { useAuthStore } from '../stores/auth';
import { useGamesStore } from '../stores/games';
import type { PreviewDeckMutation, PreviewDeckMutationVariables, DeckPreview } from '../types/generated';

const router = useRouter();
const auth = useAuthStore();
const games = useGamesStore();

const deckText = ref('');
const displayName = ref('');

const previewing = ref(false);
const previewError = ref<string | null>(null);
const previewResult = ref<DeckPreview | null>(null);

const creating = ref(false);
const createError = ref<string | null>(null);

const canSubmitPreview = computed(() => deckText.value.trim().length > 0 && !previewing.value);
const canStartTable = computed(() => Boolean(previewResult.value?.CanContinue) && !creating.value);

const totalsLine = computed(() => {
  const preview = previewResult.value;
  if (!preview) return '';
  const needsALook = preview.Unresolved.length + preview.BlockingErrors.length;
  const ready = Math.max(preview.CardCount - needsALook, 0);
  const base = `${ready} of ${preview.CardCount} card${preview.CardCount === 1 ? '' : 's'} ready`;
  if (needsALook === 0) return base;
  return `${base} - ${needsALook} need${needsALook === 1 ? 's' : ''} a quick look`;
});

async function handlePreview() {
  if (!canSubmitPreview.value) return;
  previewing.value = true;
  previewError.value = null;
  try {
    const { data } = await apolloClient.mutate<PreviewDeckMutation, PreviewDeckMutationVariables>({
      mutation: PREVIEW_DECK_MUTATION,
      variables: {
        input: {
          text: deckText.value,
          sessionID: getSessionID(),
        },
      },
    });
    previewResult.value = data?.previewDeck ?? null;
  } catch (error: unknown) {
    previewError.value = error instanceof Error ? error.message : 'We couldn’t check your deck. Please try again.';
  } finally {
    previewing.value = false;
  }
}

async function handleStartTable() {
  if (!canStartTable.value) return;
  creating.value = true;
  createError.value = null;
  try {
    if (!auth.isAuthenticated) {
      await auth.createGuestSession({
        displayName: displayName.value.trim() || undefined,
        sessionID: getSessionID(),
      });
    }

    const profile = auth.profile;
    if (!profile) {
      throw new Error('Missing profile after guest session setup');
    }

    const gameID = crypto.randomUUID();
    const payload = {
      ID: gameID,
      Turn: { Player: profile.Username, Phase: 'MAIN', Number: 1, Priority: profile.Username },
      Players: [
        {
          UserID: profile.ID,
          User: profile.Username,
          GameID: gameID,
          Life: 40,
          Commander: [],
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

    const createdID = await games.createGame(payload);
    if (!createdID) {
      throw new Error('Unable to create your table. Please try again.');
    }
    router.push({ name: 'board', params: { id: createdID } });
  } catch (error: unknown) {
    createError.value = error instanceof Error ? error.message : 'Unable to create your table. Please try again.';
  } finally {
    creating.value = false;
  }
}
</script>

<style scoped lang="scss">
.quick-start {
  max-width: 640px;
  margin: 3rem auto;
  padding: 2.5rem;
  display: grid;
  gap: 1.5rem;
}

.display-heading {
  font-size: 28px;
  font-weight: 600;
  line-height: 1.2;
  color: var(--vedh-text);
  opacity: 0.9;
}

.paste-panel {
  display: grid;
  gap: 1rem;
  padding: 1.5rem;
  border-radius: 20px;
  background: var(--vedh-panel);
  border: 1px solid var(--vedh-border);
}

.stacked {
  display: grid;
  gap: 0.5rem;
}

textarea,
input {
  padding: 0.7rem 0.9rem;
  border-radius: 10px;
  border: 1px solid var(--vedh-border);
  background: rgba(255, 244, 237, 0.05);
  color: var(--vedh-text);
  font: inherit;
}

.preview-totals {
  padding: 1rem 1.5rem;
  border-radius: 16px;
  background: var(--vedh-surface);
  border: 1px solid var(--vedh-border);
}

.actions {
  display: grid;
  gap: 0.5rem;
}

.error-text {
  color: #e5484d;
  font-size: 0.9rem;
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

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}
</style>
