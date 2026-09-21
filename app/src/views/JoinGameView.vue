<template>
  <section class="join-game">
    <h1>Join a Commander table</h1>

    <form v-if="!gameID" class="invite-entry" @submit.prevent="submitInvite">
      <label for="invite">Paste an invite link or game ID</label>
      <input id="invite" v-model.trim="inviteInput" type="text" placeholder="https://vedh.xyz/join/…" />
      <p v-if="inviteInput && !parsedID" class="hint">We couldn’t find a game ID in that invite.</p>
      <button class="primary" :disabled="!parsedID">Check table</button>
    </form>

    <p v-else-if="games.inviteLoading" class="state-card" role="status">Checking this invite…</p>

    <section v-else-if="games.inviteError === 'not_found'" class="state-card" data-testid="invite-not-found">
      <h2>This table doesn’t exist</h2>
      <p>Ask the host for a fresh invite link.</p>
    </section>

    <section v-else-if="games.inviteError === 'unavailable'" class="state-card" data-testid="invite-unavailable">
      <h2>We can’t check this invite right now</h2>
      <p>Your link is still here. Try again in a moment.</p>
      <button class="secondary" type="button" @click="loadInvite">Try again</button>
    </section>

    <template v-else-if="games.invite">
      <section class="invite-summary" data-testid="invite-summary">
        <div>
          <span class="eyebrow">{{ formatName }}</span>
          <h2>{{ games.invite.PlayerCount }} of {{ games.invite.Capacity }} seats filled</h2>
        </div>
        <ul aria-label="Players already at this table">
          <li v-for="name in games.invite.PlayerDisplayNames" :key="name">{{ name }}</li>
        </ul>
      </section>

      <section v-if="games.invite.Status === 'FINISHED'" class="state-card" data-testid="invite-finished">
        <h2>This game has finished</h2>
        <p>Ask the host to start a new table.</p>
      </section>

      <section v-else-if="games.invite.PlayerCount >= games.invite.Capacity" class="state-card" data-testid="invite-full">
        <h2>This table is full</h2>
        <p>All {{ games.invite.Capacity }} seats are already taken.</p>
      </section>

      <div v-else class="join-layout">
        <DeckImportPanel
          :initial-text="deckText"
          :initial-source-u-r-l="sourceURL"
          :session-id="sessionID"
          :context="MAGIC_COMMANDER_DECK_CONTEXT"
          @submit="handleDeckImportStarted"
          @change="handleDeckChange"
          @preview-resolved="handlePreviewResolved"
          @continue="stage = 'commander'"
        />

        <section v-if="stage === 'commander' || selectedCommanders.length" class="finish-pane">
          <CommanderReview
            :candidates="commanderCandidates"
            :selected="selectedCommanders"
            @selection-change="selectedCommanders = [...$event]"
          />

          <label v-if="!auth.isAuthenticated" class="display-name-field">
            <span>Display name (optional)</span>
            <input v-model="displayName" data-testid="join-display-name" maxlength="64" placeholder="How should the pod see you?" />
          </label>
          <p v-else class="joining-as">Joining as {{ auth.profile?.DisplayName || auth.profile?.Username }}</p>

          <button
            class="primary"
            type="button"
            data-testid="join-table"
            :disabled="!canJoin"
            @click="handleJoin"
          >
            {{ joining ? 'Joining table…' : 'Join table' }}
          </button>
        </section>
      </div>

      <section v-if="joinError" class="state-card error" data-testid="join-error" role="alert">
        <h2>The table changed while you were joining</h2>
        <p>{{ joinError }}</p>
        <button v-if="isInviteJoinable" class="secondary" type="button" @click="handleJoin">Try again</button>
      </section>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import CommanderReview from '../components/decks/CommanderReview.vue';
import DeckImportPanel, { type DeckImportChange } from '../components/decks/DeckImportPanel.vue';
import { MAGIC_COMMANDER_DECK_CONTEXT } from '../components/decks/deckImportContext';
import { lookupFormat } from '../formats/registry';
import type { CommanderPick } from '../services/commanderPartner';
import { getSessionID, track } from '../services/productEvents';
import { useAuthStore } from '../stores/auth';
import { useGamesStore } from '../stores/games';
import type { DeckPreview } from '../types/generated';

const route = useRoute();
const router = useRouter();
const games = useGamesStore();
const auth = useAuthStore();
const sessionID = getSessionID();

const inviteInput = ref('');
const deckText = ref('');
const sourceURL = ref('');
const preparedDecklist = ref('');
const previewResult = ref<DeckPreview | null>(null);
const selectedCommanders = ref<CommanderPick[]>([]);
const displayName = ref('');
const stage = ref<'import' | 'commander'>('import');
const joining = ref(false);
const joinError = ref('');
const trackedInviteIDs = new Set<string>();

const gameID = computed(() => typeof route.params.id === 'string' ? route.params.id : '');
const parsedID = computed(() => {
  const raw = inviteInput.value.trim();
  if (!raw) return '';
  const match = raw.match(/\/(?:join|games)\/([A-Za-z0-9-]+)/i);
  if (match?.[1]) return match[1];
  return /^[A-Za-z0-9-]{6,}$/.test(raw) ? raw : '';
});
const formatName = computed(() => lookupFormat(games.invite?.Format).Name);
const commanderCandidates = computed<CommanderPick[]>(() => {
  const candidates = previewResult.value?.CommanderCandidates ?? [];
  return candidates.length ? candidates : selectedCommanders.value;
});
const legalCommanderSelection = computed(() => selectedCommanders.value.length === 1 || selectedCommanders.value.length === 2);
const isInviteJoinable = computed(() => Boolean(
  games.invite
  && games.invite.Status === 'IN_PROGRESS'
  && games.invite.PlayerCount < games.invite.Capacity,
));
const canJoin = computed(() => Boolean(
  isInviteJoinable.value
  && previewResult.value?.CanContinue
  && previewResult.value.BlockingErrors.length === 0
  && legalCommanderSelection.value
  && !joining.value,
));

function submitInvite(): void {
  if (parsedID.value) void router.push({ name: 'join-game', params: { id: parsedID.value } });
}

async function loadInvite(): Promise<void> {
  if (!gameID.value) return;
  const loaded = await games.fetchGameInvite(gameID.value, sessionID);
  if (loaded && !trackedInviteIDs.has(loaded.ID)) {
    trackedInviteIDs.add(loaded.ID);
    track('invite_viewed', {}, { gameID: loaded.ID, role: 'invitee', source: 'invite' });
  }
}

function handleDeckImportStarted(): void {
  track('deck_import_started', {}, { gameID: gameID.value, role: 'invitee', source: 'invite' });
}

function handleDeckChange(change: DeckImportChange): void {
  deckText.value = change.text;
  sourceURL.value = change.sourceURL;
  preparedDecklist.value = change.decklist;
}

function handlePreviewResolved(preview: DeckPreview): void {
  previewResult.value = preview;
  joinError.value = '';
}

async function handleJoin(): Promise<void> {
  if (!canJoin.value || !gameID.value) return;
  joining.value = true;
  joinError.value = '';
  try {
    track('join_started', {}, { gameID: gameID.value, role: 'invitee', source: 'invite' });
    if (!auth.isAuthenticated) {
      await auth.createGuestSession({ displayName: displayName.value || undefined, sessionID });
    }
    const profile = auth.profile;
    if (!profile) throw new Error('join identity unavailable');

    const joinedID = await games.joinGame({
      ID: gameID.value,
      SessionID: sessionID,
      Decklist: preparedDecklist.value,
      BoardState: {
        UserID: profile.ID,
        User: profile.Username,
        GameID: gameID.value,
        Life: 40,
        Commander: selectedCommanders.value.map(({ ID, Name }) => ({ ID, Name })),
        Library: [],
        Graveyard: [],
        Exiled: [],
        Battlefield: [],
        Hand: [],
        Revealed: [],
        Controlled: [],
        Counters: [],
      },
    });
    if (!joinedID) throw new Error('join returned no game');
    await router.push({ name: 'board', params: { id: joinedID } });
  } catch {
    joinError.value = 'Your deck and commander choices are still here. Check the table and try again.';
    await loadInvite();
  } finally {
    joining.value = false;
  }
}

watch(gameID, () => void loadInvite());
onMounted(() => void loadInvite());
</script>

<style scoped lang="scss">
@use '@/styles/breakpoints' as bp;

.join-game {
  box-sizing: border-box;
  width: min(1120px, 100%);
  margin: 0 auto;
  padding: 48px 16px;
  display: grid;
  gap: 24px;
}

h1, h2, p { margin: 0; }

.invite-entry,
.finish-pane,
.display-name-field,
.state-card,
.invite-summary {
  display: grid;
  gap: 12px;
}

.invite-entry,
.state-card,
.invite-summary,
.finish-pane {
  padding: 24px;
  border: 1px solid var(--vedh-border);
  border-radius: 20px;
  background: var(--vedh-panel);
}

.invite-entry input,
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

.invite-summary ul { margin: 0; padding-left: 20px; }
.eyebrow { text-transform: uppercase; letter-spacing: 0.12em; opacity: 0.7; }
.hint, .joining-as { opacity: 0.8; }
.error { border-color: var(--vedh-danger); }
.join-layout { min-width: 0; display: grid; gap: 24px; }

button {
  border-radius: 10px;
  padding: 12px 16px;
  font: inherit;
  font-weight: 600;
  cursor: pointer;
}
button.primary { border: 0; background: var(--vedh-primary-gradient); color: var(--vedh-primary-contrast); }
button.secondary { border: 1px solid var(--vedh-border); background: var(--vedh-surface); color: var(--vedh-text); }
button:disabled { opacity: 0.5; cursor: not-allowed; }

@media (min-width: bp.$breakpoint-tablet) {
  .join-game { padding-right: 32px; padding-left: 32px; }
  .join-layout { grid-template-columns: minmax(0, 2fr) minmax(280px, 1fr); }
  .invite-summary { grid-template-columns: 1fr auto; align-items: start; }
}
</style>
