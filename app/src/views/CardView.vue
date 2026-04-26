<template>
  <section class="card-view" v-if="card">
    <header>
      <h1>{{ card.Name }}</h1>
      <p class="meta">CMC {{ card.CMC ?? '—' }} • {{ card.Types ?? '' }}</p>
    </header>
    <article>
      <p class="oracle" v-if="card.Text">{{ card.Text }}</p>
      <p class="muted" v-else>No rules text available.</p>
    </article>
  </section>
  <section v-else class="loading-state">
    <p>Loading card…</p>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { apolloClient } from '../services/apollo';
import { CARD_QUERY } from '../graphql/queriesCards';

interface CardDetail {
  ID: string;
  Name: string;
  Text?: string;
  CMC?: string;
  Types?: string;
}

const route = useRoute();
const card = ref<CardDetail | null>(null);
const cardName = computed(() => route.params.id as string | undefined);

onMounted(async () => {
  if (!cardName.value) return;
  const { data } = await apolloClient.query<{ card: CardDetail }>({
    query: CARD_QUERY,
    variables: { name: cardName.value },
  });
  card.value = data.card;
});
</script>

<style scoped lang="scss">
.card-view {
  display: grid;
  gap: 1rem;
  max-width: 640px;
}

.oracle {
  white-space: pre-wrap;
  line-height: 1.6;
}

.meta {
  color: rgba(255, 255, 255, 0.65);
}

.loading-state {
  display: grid;
  place-items: center;
  min-height: 50vh;
  color: rgba(255, 255, 255, 0.6);
}
</style>
