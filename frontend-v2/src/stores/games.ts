import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { apolloClient } from '../services/apollo';
import { CREATE_GAME_MUTATION, JOIN_GAME_MUTATION } from '../graphql/mutations';
import { GAMES_QUERY, GET_GAME_QUERY, GAME_UPDATED_SUBSCRIPTION } from '../graphql/queries';
import type { ApolloQueryResult } from '@apollo/client/core';
import type { FetchResult } from '@apollo/client/link/core';

interface PlayerSummary {
  ID?: string;
  Username: string;
}

interface TurnSummary {
  Player?: string;
  Phase?: string;
  Number?: number;
  Priority?: string;
}

interface GameSummary {
  ID: string;
  CreatedAt?: string;
  Players: PlayerSummary[];
  Turn?: TurnSummary;
  Rules?: { Name: string; Value: string }[];
}

interface BoardStateZoneCard {
  ID: string;
  Name: string;
  Types?: string; // optional, used for battlefield grouping
  CurrentZone?: string; // used to track stack owner
}

interface PlayerBoardState extends PlayerSummary {
  Boardstate?: {
    Life: number;
    Commander: BoardStateZoneCard[];
    Battlefield: BoardStateZoneCard[];
    Hand: BoardStateZoneCard[];
    Graveyard?: BoardStateZoneCard[];
    Exiled?: BoardStateZoneCard[];
    Revealed?: BoardStateZoneCard[];
    Library?: BoardStateZoneCard[];
    Controlled?: BoardStateZoneCard[];
  };
}

interface GameDetail extends GameSummary {
  Players: PlayerBoardState[];
  Stack: BoardStateZoneCard[];
}

export const useGamesStore = defineStore('games', () => {
  const games = ref<GameSummary[]>([]);
  const activeGame = ref<GameDetail | null>(null);
  const loading = ref(false);
  const errorMessage = ref<string | null>(null);
  let activeSubscription: { unsubscribe: () => void } | null = null;

  const hasActiveGame = computed(() => Boolean(activeGame.value));

  async function fetchGames(offset = 0, limit = 12) {
    loading.value = true;
    errorMessage.value = null;
    try {
      const { data }: ApolloQueryResult<{ games: GameSummary[] }> = await apolloClient.query({
        query: GAMES_QUERY,
        variables: { offset, limit },
        fetchPolicy: 'network-only',
      });
      games.value = data?.games ?? [];
    } catch (error) {
      console.error('[games] failed to fetch games', error);
      errorMessage.value = 'Unable to load games';
    } finally {
      loading.value = false;
    }
  }

  async function loadGame(gameID: string, userID?: string) {
    loading.value = true;
    errorMessage.value = null;
    try {
      const { data }: ApolloQueryResult<{ getGame: GameDetail }> = await apolloClient.query({
        query: GET_GAME_QUERY,
        variables: { gameID },
        fetchPolicy: 'network-only',
      });
      activeGame.value = data?.getGame ?? null;
      subscribeToGame(gameID, userID);
    } catch (error) {
      console.error('[games] failed to load game', error);
      errorMessage.value = 'Unable to load game';
    } finally {
      loading.value = false;
    }
  }

  function subscribeToGame(gameID: string, userID?: string) {
    if (activeSubscription) {
      activeSubscription.unsubscribe();
      activeSubscription = null;
    }
    activeSubscription = apolloClient.subscribe({
      query: GAME_UPDATED_SUBSCRIPTION,
      variables: { gameID, userID },
    }).subscribe({
      next: ({ data }: { data?: { gameUpdated?: GameDetail } }) => {
        if (data?.gameUpdated) {
          // Clone the payload to avoid assigning frozen Apollo objects into
          // our reactive store. Cloning ensures Vue's reactivity picks up
          // nested changes in templates.
          try {
            activeGame.value = JSON.parse(JSON.stringify(data.gameUpdated)) as GameDetail;
          } catch (e) {
            // Fallback to direct assignment if cloning fails
            activeGame.value = data.gameUpdated as GameDetail;
          }
        }
      },
      error: (error: unknown) => {
        console.error('[games] subscription error', error);
      },
    });
  }

  async function createGame(payload: Record<string, unknown>) {
    loading.value = true;
    try {
      const { data }: FetchResult<{ createGame: GameSummary }> = await apolloClient.mutate({
        mutation: CREATE_GAME_MUTATION,
        variables: { input: payload },
      });
      if (data?.createGame) {
        games.value = [data.createGame, ...games.value];
      }
      return data?.createGame?.ID ?? null;
    } finally {
      loading.value = false;
    }
  }

  async function joinGame(payload: Record<string, unknown>) {
    loading.value = true;
    try {
      const { data }: FetchResult<{ joinGame: GameSummary }> = await apolloClient.mutate({
        mutation: JOIN_GAME_MUTATION,
        variables: { input: payload },
      });
      if (data?.joinGame) {
        activeGame.value = data.joinGame as unknown as GameDetail;
      }
      return data?.joinGame?.ID ?? null;
    } finally {
      loading.value = false;
    }
  }

  function clearActiveGame() {
    activeGame.value = null;
    if (activeSubscription) {
      activeSubscription.unsubscribe();
      activeSubscription = null;
    }
  }

  return {
    games,
    activeGame,
    loading,
    errorMessage,
    hasActiveGame,
    fetchGames,
    loadGame,
    subscribeToGame,
    createGame,
    joinGame,
    clearActiveGame,
  };
});
