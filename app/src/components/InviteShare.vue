<template>
  <div class="invite-share">
    <button class="invite-action" type="button" data-testid="share-invite" @click="shareInvite">Invite players</button>
    <p v-if="feedback" class="share-feedback" role="status">{{ feedback }}</p>
    <label v-if="showManual" class="manual-invite">
      <span>Copy this invite link manually</span>
      <input :value="inviteURL" readonly @focus="selectManualInvite" />
    </label>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { track } from '../services/productEvents';

const props = defineProps<{
  gameID: string;
  role: 'host' | 'invitee';
}>();

const feedback = ref('');
const showManual = ref(false);
const inviteURL = computed(() => {
  const origin = typeof window !== 'undefined' ? window.location.origin : '';
  return `${origin}/join/${props.gameID}`;
});

async function shareInvite(): Promise<void> {
  feedback.value = '';
  showManual.value = false;
  let shareMethod = '';

  if (typeof navigator !== 'undefined' && typeof navigator.share === 'function') {
    try {
      await navigator.share({ title: 'Join my vEDH table', url: inviteURL.value });
      shareMethod = 'native_share';
    } catch {
      // Continue to clipboard fallback.
    }
  }

  if (!shareMethod && typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(inviteURL.value);
      shareMethod = 'clipboard';
    } catch {
      // The selectable field below is the final fallback.
    }
  }

  if (!shareMethod) {
    showManual.value = true;
    feedback.value = 'Automatic sharing was blocked.';
    return;
  }

  feedback.value = shareMethod === 'native_share' ? 'Invite shared.' : 'Invite link copied.';
  track('invite_copied', { share_method: shareMethod }, {
    gameID: props.gameID,
    role: props.role,
    source: 'board',
  });
}

function selectManualInvite(event: FocusEvent): void {
  (event.target as HTMLInputElement | null)?.select();
}
</script>

<style scoped>
.invite-share { position: relative; }
.invite-action {
  border: 0;
  border-radius: 10px;
  padding: 0.7rem 1rem;
  background: var(--vedh-primary-gradient);
  color: var(--vedh-primary-contrast);
  font: inherit;
  font-weight: 700;
  cursor: pointer;
}
.share-feedback,
.manual-invite {
  position: absolute;
  z-index: 20;
  top: calc(100% + 0.5rem);
  right: 0;
  width: min(360px, 80vw);
  box-sizing: border-box;
  margin: 0;
  border: 1px solid var(--vedh-border);
  border-radius: 12px;
  padding: 0.75rem 1rem;
  background: var(--vedh-panel-strong);
}
.manual-invite { display: grid; gap: 0.5rem; }
.manual-invite input {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid var(--vedh-border);
  border-radius: 8px;
  padding: 0.65rem 0.75rem;
  background: rgba(255, 244, 237, 0.05);
  color: var(--vedh-text);
}
</style>
