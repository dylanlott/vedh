<template>
  <aside class="mobile-heads-up" data-testid="mobile-heads-up" role="note">
    This board is built for tablets and larger. You can finish importing your deck here — the board itself may feel tight on a phone.
  </aside>
</template>

<style scoped lang="scss">
@use '@/styles/breakpoints' as bp;

.mobile-heads-up {
  box-sizing: border-box;
  width: 100%;
  padding: 8px 16px;
  border: 1px solid var(--vedh-border);
  border-radius: 10px;
  background: var(--vedh-surface);
  color: var(--vedh-text);
  font-size: 14px;
  font-weight: 600;
  line-height: 1.4;
  overflow-wrap: anywhere;
}

@media (min-width: bp.$breakpoint-tablet) {
  .mobile-heads-up { display: none; }
}
</style>
