import { defineStore } from 'pinia';
import { ref } from 'vue';
import { apolloClient } from '../services/apollo';
import { CARD_QUERY, CARD_SEARCH_QUERY } from '../graphql/queriesCards';

interface CardSummary {
  ID: string;
  Name: string;
  Types?: string;
}

export const useCardsStore = defineStore('cards', () => {
  const searchResults = ref<CardSummary[]>([]);
  const loading = ref(false);

  async function search(name: string) {
    if (!name) {
      searchResults.value = [];
      return;
    }
    loading.value = true;
    try {
      const { data } = await apolloClient.query<{ search: CardSummary[] }>({
        query: CARD_SEARCH_QUERY,
        variables: { name },
      });
      searchResults.value = data.search ?? [];
    } finally {
      loading.value = false;
    }
  }

  async function fetchCard(name: string) {
    const { data } = await apolloClient.query<{ card: CardSummary }>({
      query: CARD_QUERY,
      variables: { name },
    });
    return data.card;
  }

  return {
    searchResults,
    loading,
    search,
    fetchCard,
  };
});
