<template>
  <aside v-if="!dismissed && !claimed" class="claim-account" aria-labelledby="claim-account-title">
    <button class="dismiss" type="button" aria-label="Dismiss save games prompt" @click="dismiss">×</button>
    <div class="copy">
      <span class="eyebrow">Keep this table</span>
      <h2 id="claim-account-title">Save your games</h2>
      <p>Choose a login without leaving the board. Your current game stays exactly where it is.</p>
    </div>
    <form @submit.prevent="submit">
      <label>
        Username
        <input v-model="username" name="username" autocomplete="username" required />
      </label>
      <label>
        Password
        <input v-model="password" name="password" type="password" autocomplete="new-password" required />
      </label>
      <p v-if="errorMessage" class="error" role="alert">{{ errorMessage }}</p>
      <button class="primary" type="submit" :disabled="submitting">
        {{ submitting ? 'Saving…' : 'Save my games' }}
      </button>
    </form>
  </aside>
  <p v-else-if="claimed" class="claim-success" role="status">Games saved. You can keep playing.</p>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useAuthStore } from '../../stores/auth';
import { getSessionID, track } from '../../services/productEvents';

const props = defineProps<{
  gameID: string;
  role: 'host' | 'invitee';
}>();

const auth = useAuthStore();
const username = ref(auth.profile?.DisplayName?.trim() || auth.profile?.Username || '');
const password = ref('');
const submitting = ref(false);
const claimed = ref(false);
const dismissalKey = `edhgo/claim-dismissed/${props.gameID}`;
const dismissed = ref(readDismissed());
const errorMessage = ref('');

function readDismissed(): boolean {
  try {
    return sessionStorage.getItem(dismissalKey) === 'true';
  } catch {
    return false;
  }
}

function dismiss(): void {
  dismissed.value = true;
  try {
    sessionStorage.setItem(dismissalKey, 'true');
  } catch {
    // A disabled storage surface should not make dismissal affect gameplay.
  }
}

async function submit(): Promise<void> {
  if (submitting.value) return;
  errorMessage.value = '';
  const trimmedUsername = username.value.trim();
  if (!trimmedUsername || !password.value) {
    errorMessage.value = 'Choose a username and password.';
    return;
  }

  submitting.value = true;
  const sessionID = getSessionID();
  track('account_claim_started', {}, {
    gameID: props.gameID,
    role: props.role,
    source: 'board',
  });
  try {
    await auth.claimGuestAccount({
      username: trimmedUsername,
      password: password.value,
      sessionID,
    });
    password.value = '';
    claimed.value = true;
  } catch (error: unknown) {
    errorMessage.value = error instanceof Error ? error.message : 'Could not save this account. Try again.';
  } finally {
    submitting.value = false;
  }
}
</script>

<style scoped lang="scss">
.claim-account,
.claim-success {
  position: relative;
  margin: 0.75rem 0;
  padding: 1rem;
  border: 1px solid var(--vedh-border);
  border-radius: 0.75rem;
  background: rgba(35, 31, 29, 0.94);
}

.claim-account {
  display: grid;
  grid-template-columns: minmax(12rem, 1fr) minmax(16rem, 1.25fr);
  gap: 1rem;
}

.copy,
form,
label {
  display: grid;
  gap: 0.45rem;
}

h2,
p {
  margin: 0;
}

.eyebrow {
  color: var(--vedh-primary);
  font-size: 0.72rem;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

input {
  min-width: 0;
  padding: 0.65rem 0.75rem;
  border: 1px solid var(--vedh-border);
  border-radius: 0.45rem;
  color: var(--vedh-text);
  background: rgba(15, 13, 12, 0.62);
}

.dismiss {
  position: absolute;
  top: 0.4rem;
  right: 0.55rem;
  border: 0;
  color: var(--vedh-muted);
  background: transparent;
  font-size: 1.35rem;
  cursor: pointer;
}

.error {
  color: #ffb4a8;
}

@media (max-width: 720px) {
  .claim-account {
    grid-template-columns: 1fr;
  }
}
</style>
