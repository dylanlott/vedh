import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';

const apollo = vi.hoisted(() => ({
  query: vi.fn(),
  subscribe: vi.fn(),
  realtimeListener: undefined as undefined | ((state: string) => void),
  subscriptionHandlers: undefined as undefined | { next: (value: unknown) => void; error: (error: unknown) => void },
}));

vi.mock('../src/services/apollo', () => ({
  apolloClient: {
    query: apollo.query,
    mutate: vi.fn(),
    subscribe: apollo.subscribe,
  },
  onRealtimeConnectionState: vi.fn((listener) => {
    apollo.realtimeListener = listener;
    listener('idle');
    return () => undefined;
  }),
}));

import { useGamesStore } from '../src/stores/games';

const GAME = {
  ID: 'game-1',
  Players: [{
    ID: 'user-1',
    Username: 'Player',
    Boardstate: { Life: 40, Commander: [], Battlefield: [], Hand: [], Library: [] },
  }],
  Stack: [],
};

beforeEach(() => {
  vi.useFakeTimers();
  setActivePinia(createPinia());
  apollo.query.mockReset();
  apollo.subscribe.mockReset();
  apollo.query.mockResolvedValue({ data: { getGame: GAME } });
  apollo.subscribe.mockReturnValue({
    subscribe: vi.fn((handlers) => {
      apollo.subscriptionHandlers = handlers;
      return { unsubscribe: vi.fn() };
    }),
  });
});

describe('games store board readiness', () => {
  it('becomes ready only after a usable game and realtime connection exist', async () => {
    const games = useGamesStore();
    const loading = games.loadGame('game-1', 'user-1');
    expect(games.boardConnectionState).toBe('loading');
    await loading;
    expect(games.boardConnectionState).toBe('loading');

    apollo.realtimeListener?.('connected');
    expect(games.boardConnectionState).toBe('ready');
  });

  it('enters documented degraded polling after subscription failure', async () => {
    const games = useGamesStore();
    await games.loadGame('game-1', 'user-1');
    apollo.subscriptionHandlers?.error(new Error('socket down'));

    expect(games.boardConnectionState).toBe('degraded');
    await vi.advanceTimersByTimeAsync(3_000);
    expect(apollo.query).toHaveBeenCalledTimes(2);
    expect(games.boardConnectionState).toBe('degraded');
  });

  it('bounds degraded polling and exposes a terminal failure', async () => {
    const games = useGamesStore();
    await games.loadGame('game-1', 'user-1');
    apollo.subscriptionHandlers?.error(new Error('socket down'));

    await vi.advanceTimersByTimeAsync(30_000);
    expect(games.boardConnectionState).toBe('failed');
  });

  it('reconnects from degraded mode and returns to ready', async () => {
    const games = useGamesStore();
    await games.loadGame('game-1', 'user-1');
    apollo.subscriptionHandlers?.error(new Error('socket down'));
    expect(games.boardConnectionState).toBe('degraded');

    games.reconnectGame();
    expect(apollo.subscribe).toHaveBeenCalledTimes(2);
    expect(games.boardConnectionState).toBe('reconnecting');
    apollo.realtimeListener?.('connected');
    expect(games.boardConnectionState).toBe('ready');
  });
});
