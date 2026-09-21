import { beforeEach, describe, expect, it, vi } from 'vitest';
import { flushPromises, mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';

const pushMock = vi.fn();
vi.mock('vue-router', () => ({
  useRoute: () => ({ params: { id: 'game-1' } }),
  useRouter: () => ({ push: pushMock }),
}));

const productEvents = vi.hoisted(() => ({
  getSessionID: vi.fn(() => 'join-session'),
  track: vi.fn(),
}));
vi.mock('../src/services/productEvents', () => productEvents);

import CommanderReview from '../src/components/decks/CommanderReview.vue';
import DeckImportPanel from '../src/components/decks/DeckImportPanel.vue';
import { MAGIC_COMMANDER_DECK_CONTEXT } from '../src/components/decks/deckImportContext';
import JoinGameView from '../src/views/JoinGameView.vue';
import { useAuthStore } from '../src/stores/auth';
import { useGamesStore, type GameInvite } from '../src/stores/games';

const VALID_INVITE: GameInvite = {
  ID: 'game-1',
  Format: 'EDH',
  Status: 'IN_PROGRESS',
  PlayerDisplayNames: ['Host'],
  PlayerCount: 1,
  Capacity: 4,
  CreatedAt: '2026-09-21T00:00:00Z',
};

const READY_PREVIEW = {
  SourceType: 'PASTE',
  CardCount: 100,
  Entries: [],
  CommanderCandidates: [{ ID: 'commander-1', Name: 'Atraxa', Text: '' }],
  Unresolved: [],
  Warnings: [],
  BlockingErrors: [],
  CanContinue: true,
};

function mockInvite(invite: GameInvite | null = VALID_INVITE) {
  const games = useGamesStore();
  vi.spyOn(games, 'fetchGameInvite').mockImplementation(async () => {
    games.invite = invite;
    games.inviteError = invite ? null : 'not_found';
    return invite;
  });
  return games;
}

async function prepareJoin(wrapper: ReturnType<typeof mount>) {
  const panel = wrapper.getComponent(DeckImportPanel);
  panel.vm.$emit('change', {
    text: '1 Sol Ring\n99 Island',
    sourceURL: '',
    corrections: {},
    decklist: '1 Sol Ring\n99 Island',
  });
  panel.vm.$emit('preview-resolved', READY_PREVIEW);
  panel.vm.$emit('continue');
  await flushPromises();
  wrapper.getComponent(CommanderReview).vm.$emit('selection-change', [{ ID: 'commander-1', Name: 'Atraxa' }]);
  await flushPromises();
}

beforeEach(() => {
  localStorage.clear();
  localStorage.setItem('edhgo/auth', JSON.stringify({ ID: 'user-2', Username: 'Joiner', Token: 'token' }));
  setActivePinia(createPinia());
  pushMock.mockReset();
  productEvents.track.mockReset();
});

describe('JoinGame public activation flow', () => {
  it('loads the safe invite before rendering the shared deck import', async () => {
    const games = mockInvite();
    const wrapper = mount(JoinGameView);
    await flushPromises();

    expect(games.fetchGameInvite).toHaveBeenCalledWith('game-1', 'join-session');
    expect(wrapper.get('[data-testid="invite-summary"]').text()).toContain('1 of 4 seats filled');
    expect(wrapper.get('[data-testid="invite-summary"]').text()).toContain('Host');
    const panel = wrapper.getComponent(DeckImportPanel);
    expect(panel.props('context')).toEqual(MAGIC_COMMANDER_DECK_CONTEXT);
    expect(wrapper.text()).not.toContain('quantity,name per line');
    expect(productEvents.track).toHaveBeenCalledWith('invite_viewed', {}, {
      gameID: 'game-1', role: 'invitee', source: 'invite',
    });
  });

  it('submits the canonical deck through the existing join mutation for an authenticated player', async () => {
    const games = mockInvite();
    const joinGame = vi.spyOn(games, 'joinGame').mockResolvedValue('game-1');
    const wrapper = mount(JoinGameView);
    await flushPromises();
    await prepareJoin(wrapper);

    await wrapper.get('[data-testid="join-table"]').trigger('click');
    await flushPromises();

    expect(joinGame).toHaveBeenCalledWith(expect.objectContaining({
      ID: 'game-1',
      SessionID: 'join-session',
      Decklist: '1 Sol Ring\n99 Island',
      BoardState: expect.objectContaining({
        UserID: 'user-2',
        User: 'Joiner',
        Commander: [{ ID: 'commander-1', Name: 'Atraxa' }],
      }),
    }));
    expect(pushMock).toHaveBeenCalledWith({ name: 'board', params: { id: 'game-1' } });
  });

  it('creates a guest only after preview and commander review are complete', async () => {
    localStorage.clear();
    setActivePinia(createPinia());
    const games = mockInvite();
    vi.spyOn(games, 'joinGame').mockResolvedValue('game-1');
    const auth = useAuthStore();
    const createGuest = vi.spyOn(auth, 'createGuestSession').mockImplementation(async () => {
      auth.profile = { ID: 'guest-id', Username: 'BraveSliver', Token: 'guest-token', IsGuest: true };
      return auth.profile;
    });
    const wrapper = mount(JoinGameView);
    await flushPromises();

    expect(createGuest).not.toHaveBeenCalled();
    await prepareJoin(wrapper);
    expect(createGuest).not.toHaveBeenCalled();
    await wrapper.get('[data-testid="join-display-name"]').setValue('Dylan');
    await wrapper.get('[data-testid="join-table"]').trigger('click');
    await flushPromises();

    expect(createGuest).toHaveBeenCalledWith({ displayName: 'Dylan', sessionID: 'join-session' });
  });

  it.each([
    { state: 'finished', invite: { ...VALID_INVITE, Status: 'FINISHED' as const }, testID: 'invite-finished' },
    { state: 'full', invite: { ...VALID_INVITE, PlayerCount: 4 }, testID: 'invite-full' },
  ])('shows the $state state before requesting a deck', async ({ invite, testID }) => {
    mockInvite(invite);
    const wrapper = mount(JoinGameView);
    await flushPromises();

    expect(wrapper.get(`[data-testid="${testID}"]`).exists()).toBe(true);
    expect(wrapper.findComponent(DeckImportPanel).exists()).toBe(false);
  });

  it('shows a missing invite before requesting a deck', async () => {
    mockInvite(null);
    const wrapper = mount(JoinGameView);
    await flushPromises();

    expect(wrapper.get('[data-testid="invite-not-found"]').exists()).toBe(true);
    expect(wrapper.findComponent(DeckImportPanel).exists()).toBe(false);
  });

  it('keeps the prepared deck and commander when a join race is recoverable', async () => {
    const games = mockInvite();
    vi.spyOn(games, 'joinGame').mockRejectedValue(new Error('raw server detail'));
    const wrapper = mount(JoinGameView);
    await flushPromises();
    await prepareJoin(wrapper);

    await wrapper.get('[data-testid="join-table"]').trigger('click');
    await flushPromises();

    expect(wrapper.get('[data-testid="join-error"]').text()).not.toContain('raw server detail');
    expect(wrapper.getComponent(DeckImportPanel).props('initialText')).toBe('1 Sol Ring\n99 Island');
    expect(wrapper.getComponent(CommanderReview).props('selected')).toEqual([{ ID: 'commander-1', Name: 'Atraxa' }]);
  });
});
