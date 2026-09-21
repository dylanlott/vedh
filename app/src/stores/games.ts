import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { apolloClient, onRealtimeConnectionState, type RealtimeConnectionState } from '../services/apollo';
import { CREATE_GAME_MUTATION, JOIN_GAME_MUTATION } from '../graphql/mutations';
import { GAMES_QUERY, GET_GAME_QUERY, GAME_UPDATED_SUBSCRIPTION, FORMATS_QUERY, GAME_INVITE_QUERY } from '../graphql/queries';
import { formats as localFormats, lookupFormat, type GameFormat } from '../formats/registry';
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
  Status?: string;
  Result?: string;
  WinnerIDs?: string[];
  WinCondition?: string;
  PendingWinClaim?: {
    ClaimedBy: string;
    Condition?: string;
    Remaining: string[];
  };
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

export interface GameInvite {
  ID: string;
  Format: string;
  Status: 'IN_PROGRESS' | 'FINISHED';
  PlayerDisplayNames: string[];
  PlayerCount: number;
  Capacity: number;
  CreatedAt: string;
}

export type BoardConnectionState = 'idle' | 'loading' | 'ready' | 'degraded' | 'reconnecting' | 'failed';

const REALTIME_CONNECT_TIMEOUT_MS = 2_000;
const DEGRADED_POLL_INTERVAL_MS = 3_000;
const MAX_DEGRADED_POLLS = 10;

export const useGamesStore = defineStore('games', () => {
  const games = ref<GameSummary[]>([]);
  const activeGame = ref<GameDetail | null>(null);
  const formats = ref<GameFormat[]>([...localFormats]);
  const loading = ref(false);
  const errorMessage = ref<string | null>(null);
  const invite = ref<GameInvite | null>(null);
  const inviteLoading = ref(false);
  const inviteError = ref<'not_found' | 'unavailable' | null>(null);
  const boardConnectionState = ref<BoardConnectionState>('idle');
  let activeSubscription: { unsubscribe: () => void } | null = null;
  let activeGameID = '';
  let activeUserID = '';
  let realtimeState: RealtimeConnectionState = 'unavailable';
  let connectTimer: number | null = null;
  let pollingTimer: number | null = null;
  let degradedPolls = 0;

  const hasActiveGame = computed(() => Boolean(activeGame.value));

  function hasCurrentPlayerBoard(): boolean {
    return Boolean(activeGame.value?.Players.some((player) => player.ID === activeUserID && player.Boardstate));
  }

  function clearConnectTimer(): void {
    if (connectTimer !== null) window.clearTimeout(connectTimer);
    connectTimer = null;
  }

  function stopPolling(): void {
    if (pollingTimer !== null) window.clearInterval(pollingTimer);
    pollingTimer = null;
    degradedPolls = 0;
  }

  async function fetchActiveGameSnapshot(): Promise<boolean> {
    if (!activeGameID) return false;
    try {
      const { data }: ApolloQueryResult<{ getGame: GameDetail }> = await apolloClient.query({
        query: GET_GAME_QUERY,
        variables: { gameID: activeGameID },
        fetchPolicy: 'network-only',
      });
      activeGame.value = data?.getGame ?? null;
      return hasCurrentPlayerBoard();
    } catch (error) {
      console.error('[games] degraded poll failed', error);
      return false;
    }
  }

  function beginDegradedPolling(): void {
    clearConnectTimer();
    if (!activeGame.value || !hasCurrentPlayerBoard()) {
      boardConnectionState.value = 'failed';
      return;
    }
    boardConnectionState.value = 'degraded';
    if (pollingTimer !== null) return;
    degradedPolls = 0;
    pollingTimer = window.setInterval(() => {
      degradedPolls += 1;
      void fetchActiveGameSnapshot().then((usable) => {
        if (usable) boardConnectionState.value = 'degraded';
        if (degradedPolls >= MAX_DEGRADED_POLLS) {
          stopPolling();
          boardConnectionState.value = 'failed';
        }
      });
    }, DEGRADED_POLL_INTERVAL_MS);
  }

  function markRealtimeReady(): void {
    if (!activeGame.value || !hasCurrentPlayerBoard()) return;
    clearConnectTimer();
    stopPolling();
    boardConnectionState.value = 'ready';
  }

  onRealtimeConnectionState((state) => {
    realtimeState = state;
    if (!activeSubscription) return;
    if (state === 'connected') {
      markRealtimeReady();
    } else if (state === 'closed' || state === 'error' || state === 'unavailable') {
      beginDegradedPolling();
    } else if (state === 'connecting' && boardConnectionState.value !== 'loading') {
      boardConnectionState.value = 'reconnecting';
    }
  });

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

  async function fetchFormats() {
	try {
	  const { data }: ApolloQueryResult<{ formats: GameFormat[] }> = await apolloClient.query({
	    query: FORMATS_QUERY,
	    fetchPolicy: 'network-only',
	  });
	  formats.value = data?.formats?.length ? data.formats : [...localFormats];
	} catch {
	  formats.value = [...localFormats];
	}
  }

  function resolveGameFormat(rules?: { Name: string; Value: string }[] | null) {
	return lookupFormat(rules?.find(rule => rule.Name === 'format')?.Value);
  }

  async function loadGame(gameID: string, userID?: string) {
    loading.value = true;
    errorMessage.value = null;
    boardConnectionState.value = 'loading';
    activeGameID = gameID;
    activeUserID = userID ?? '';
    try {
      const { data }: ApolloQueryResult<{ getGame: GameDetail }> = await apolloClient.query({
        query: GET_GAME_QUERY,
        variables: { gameID },
        fetchPolicy: 'network-only',
      });
      activeGame.value = data?.getGame ?? null;
      if (!activeGame.value || !activeUserID || !hasCurrentPlayerBoard()) {
        boardConnectionState.value = 'failed';
        return;
      }
      subscribeToGame(gameID, userID);
    } catch (error) {
      console.error('[games] failed to load game', error);
      errorMessage.value = 'Unable to load game';
      boardConnectionState.value = 'failed';
    } finally {
      loading.value = false;
    }
  }

  function subscribeToGame(gameID: string, userID?: string) {
    if (activeSubscription) {
      activeSubscription.unsubscribe();
      activeSubscription = null;
    }
    clearConnectTimer();
    stopPolling();
    activeGameID = gameID;
    activeUserID = userID ?? activeUserID;
    if (!activeUserID) return;
    boardConnectionState.value = boardConnectionState.value === 'loading' ? 'loading' : 'reconnecting';
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
          markRealtimeReady();
        }
      },
      error: (error: unknown) => {
        console.error('[games] subscription error', error);
        beginDegradedPolling();
      },
    });

    if (realtimeState === 'connected') {
      markRealtimeReady();
    } else {
      connectTimer = window.setTimeout(beginDegradedPolling, REALTIME_CONNECT_TIMEOUT_MS);
    }
  }

  function reconnectGame(): void {
    if (!activeGameID || !activeUserID) {
      boardConnectionState.value = 'failed';
      return;
    }
    subscribeToGame(activeGameID, activeUserID);
  }

  async function fetchGameInvite(gameID: string, sessionID: string): Promise<GameInvite | null> {
    inviteLoading.value = true;
    inviteError.value = null;
    invite.value = null;
    try {
      const { data }: ApolloQueryResult<{ gameInvite: GameInvite | null }> = await apolloClient.query({
        query: GAME_INVITE_QUERY,
        variables: { gameID, sessionID },
        fetchPolicy: 'network-only',
      });
      invite.value = data?.gameInvite ?? null;
      if (!invite.value) inviteError.value = 'not_found';
      return invite.value;
    } catch (error) {
      console.error('[games] failed to load invite', error);
      inviteError.value = 'unavailable';
      return null;
    } finally {
      inviteLoading.value = false;
    }
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
    clearConnectTimer();
    stopPolling();
    activeGameID = '';
    activeUserID = '';
    boardConnectionState.value = 'idle';
  }

  return {
    games,
    activeGame,
    loading,
    errorMessage,
    invite,
    inviteLoading,
    inviteError,
    boardConnectionState,
    hasActiveGame,
    formats,
    fetchGames,
    fetchFormats,
    loadGame,
    subscribeToGame,
    reconnectGame,
    fetchGameInvite,
    createGame,
    joinGame,
    clearActiveGame,
    resolveGameFormat,
  };
});
