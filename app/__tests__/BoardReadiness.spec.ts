import { beforeEach, describe, expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { nextTick, reactive } from 'vue';

const gamesContainer = vi.hoisted(() => ({ state: undefined as unknown }));
vi.mock('../src/stores/games', () => ({ useGamesStore: () => gamesContainer.state }));

const authStore = vi.hoisted(() => ({ profile: { ID: 'user-1', Username: 'Player', Token: 'token' } }));
vi.mock('../src/stores/auth', () => ({ useAuthStore: () => authStore }));

vi.mock('vue-router', () => ({
  useRoute: () => ({ params: { id: 'game-1' } }),
  useRouter: () => ({ push: vi.fn() }),
}));

const productEvents = vi.hoisted(() => ({ track: vi.fn() }));
vi.mock('../src/services/productEvents', () => productEvents);

import BoardView from '../src/views/BoardView.vue';

const GAME = {
  ID: 'game-1',
  CreatedAt: '2026-09-21T00:00:00Z',
  Status: 'IN_PROGRESS',
  Players: [{
    ID: 'user-1',
    Username: 'Player',
    Boardstate: {
      Life: 40,
      Commander: [], Battlefield: [], Hand: [], Graveyard: [], Exiled: [], Revealed: [], Library: [], Controlled: [],
    },
  }],
  Stack: [],
  Turn: { Player: 'Player', Priority: 'Player', Phase: 'MAIN', Number: 1 },
  Rules: [],
};

beforeEach(() => {
  productEvents.track.mockReset();
  gamesContainer.state = reactive({
    activeGame: null,
    boardConnectionState: 'idle',
    loadGame: vi.fn().mockResolvedValue(undefined),
    subscribeToGame: vi.fn(),
    reconnectGame: vi.fn(),
    clearActiveGame: vi.fn(),
  });
});

describe('BoardView readiness telemetry', () => {
  it('does not emit on route navigation or query data alone, then emits once after degraded polling is active', async () => {
    const wrapper = mount(BoardView, {
      global: { stubs: { Card: true, InviteShare: true } },
    });
    await nextTick();

    expect(productEvents.track).not.toHaveBeenCalled();
    const games = gamesContainer.state as { activeGame: typeof GAME | null; boardConnectionState: string };
    games.activeGame = GAME;
    games.boardConnectionState = 'loading';
    await nextTick();
    expect(productEvents.track).not.toHaveBeenCalled();

    games.boardConnectionState = 'degraded';
    await nextTick();
    expect(productEvents.track).toHaveBeenCalledTimes(1);
    expect(productEvents.track).toHaveBeenCalledWith('board_ready', {}, expect.objectContaining({
      gameID: 'game-1', role: 'host', source: 'board', outcome: 'degraded',
    }));

    games.boardConnectionState = 'ready';
    await nextTick();
    expect(productEvents.track).toHaveBeenCalledTimes(1);
    wrapper.unmount();
  });
});
