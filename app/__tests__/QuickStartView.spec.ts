import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';

// The Apollo client is mocked at the module boundary so this suite never
// makes a real network call, mirroring app/__tests__/productEvents.spec.ts's
// convention.
vi.mock('../src/services/apollo', () => ({
  apolloClient: { mutate: vi.fn() },
}));

const pushMock = vi.fn();
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
}));

import { apolloClient } from '../src/services/apollo';
import QuickStartView from '../src/views/QuickStartView.vue';
import { useAuthStore } from '../src/stores/auth';

type MutateMock = ReturnType<typeof vi.fn>;

function mutate(): MutateMock {
  return apolloClient.mutate as unknown as MutateMock;
}

const READY_PREVIEW = {
  SourceType: 'PASTE',
  CardCount: 100,
  Entries: [],
  CommanderCandidates: [],
  Unresolved: [],
  Warnings: [],
  CanContinue: true,
  BlockingErrors: [],
};

const GUEST_SESSION_RESULT = {
  guestSession: {
    ID: 'guest-1',
    Username: 'Brave Sliver',
    DisplayName: null,
    IsGuest: true,
    Token: 'guest-jwt',
    GuestCredential: 'guest-1.secret',
  },
};

const CREATE_GAME_RESULT = { createGame: { ID: 'game-1' } };

beforeEach(() => {
  localStorage.clear();
  setActivePinia(createPinia());
  mutate().mockReset();
  pushMock.mockReset();
});

describe('QuickStartView', () => {
  it('renders the paste surface for a logged-out visitor with no navigation to login/signup', () => {
    const wrapper = mount(QuickStartView);
    expect(wrapper.text()).toContain('Paste your decklist');
    expect(pushMock).not.toHaveBeenCalled();
  });

  it('previewDeck is called once with { text, sessionID } and the totals line renders', async () => {
    mutate().mockResolvedValueOnce({ data: { previewDeck: READY_PREVIEW } });

    const wrapper = mount(QuickStartView);
    await wrapper.find('textarea').setValue('1, Sol Ring');
    await wrapper.find('form').trigger('submit');
    await flushPromises();

    expect(mutate()).toHaveBeenCalledTimes(1);
    const call = mutate().mock.calls[0][0];
    expect(call.variables.input.text).toBe('1, Sol Ring');
    expect(typeof call.variables.input.sessionID).toBe('string');
    expect(call.variables.input.sessionID.length).toBeGreaterThan(0);

    expect(wrapper.text()).toContain('100 of 100 cards ready');
  });

  it('with CanContinue true: calls guestSession once, then createGame once, then navigates to the board', async () => {
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: READY_PREVIEW } })
      .mockResolvedValueOnce({ data: GUEST_SESSION_RESULT })
      .mockResolvedValueOnce({ data: CREATE_GAME_RESULT });

    const wrapper = mount(QuickStartView);
    await wrapper.find('textarea').setValue('1, Sol Ring');
    await wrapper.find('form').trigger('submit');
    await flushPromises();

    await wrapper.find('[data-testid="start-table"]').trigger('click');
    await flushPromises();

    expect(mutate()).toHaveBeenCalledTimes(3);
    const guestCall = mutate().mock.calls[1][0];
    expect(guestCall.variables.sessionID).toBeTruthy();
    const createGameCall = mutate().mock.calls[2][0];
    expect(createGameCall.variables.input.Players[0].User).toBe('Brave Sliver');

    expect(pushMock).toHaveBeenCalledWith({ name: 'board', params: { id: 'game-1' } });
  });

  it('with an authenticated auth store: the same path runs and guestSession is never called', async () => {
    const auth = useAuthStore();
    auth.profile = { ID: 'real-user-1', Username: 'RealUser', Token: 'real-jwt' };

    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: READY_PREVIEW } })
      .mockResolvedValueOnce({ data: CREATE_GAME_RESULT });

    const wrapper = mount(QuickStartView);
    await wrapper.find('textarea').setValue('1, Sol Ring');
    await wrapper.find('form').trigger('submit');
    await flushPromises();

    await wrapper.find('[data-testid="start-table"]').trigger('click');
    await flushPromises();

    expect(mutate()).toHaveBeenCalledTimes(2);
    const createGameCall = mutate().mock.calls[1][0];
    expect(createGameCall.variables.input.Players[0].User).toBe('RealUser');
    expect(pushMock).toHaveBeenCalledWith({ name: 'board', params: { id: 'game-1' } });
  });

  it('an empty paste leaves the submit control disabled and previewDeck uncalled', async () => {
    const wrapper = mount(QuickStartView);
    const submit = wrapper.find('[data-testid="preview-submit"]');
    expect(submit.attributes('disabled')).toBeDefined();

    await wrapper.find('form').trigger('submit');
    await flushPromises();
    expect(mutate()).not.toHaveBeenCalled();
  });

  it('submitting the same text twice results in one guest and one game', async () => {
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: READY_PREVIEW } })
      .mockResolvedValueOnce({ data: { previewDeck: READY_PREVIEW } })
      .mockResolvedValueOnce({ data: GUEST_SESSION_RESULT })
      .mockResolvedValueOnce({ data: CREATE_GAME_RESULT });

    const wrapper = mount(QuickStartView);
    await wrapper.find('textarea').setValue('1, Sol Ring');
    await wrapper.find('form').trigger('submit');
    await flushPromises();
    // Submitting the identical paste again (e.g. a second click before
    // noticing the first request already resolved) re-runs previewDeck but
    // must not multiply anything downstream.
    await wrapper.find('form').trigger('submit');
    await flushPromises();

    await wrapper.find('[data-testid="start-table"]').trigger('click');
    await flushPromises();

    expect(mutate()).toHaveBeenCalledTimes(4);
    const guestSessionCalls = mutate().mock.calls.filter((call) => 'displayName' in (call[0].variables ?? {}));
    const createGameCalls = mutate().mock.calls.filter((call) => call[0].variables?.input?.Players);
    expect(guestSessionCalls).toHaveLength(1);
    expect(createGameCalls).toHaveLength(1);
  });
});
