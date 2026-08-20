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
  CommanderCandidates: [{ ID: 'commander-1', Name: 'Atraxa', Text: '' }],
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
  sessionStorage.clear();
  setActivePinia(createPinia());
  mutate().mockReset();
  pushMock.mockReset();
});

async function chooseCommander(wrapper: ReturnType<typeof mount>) {
  await wrapper.get('[data-testid="continue-to-commanders"]').trigger('click');
  await wrapper.get('[data-testid="commander-candidate"]').trigger('click');
}

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

    await chooseCommander(wrapper);
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

    await chooseCommander(wrapper);
    await wrapper.find('[data-testid="start-table"]').trigger('click');
    await flushPromises();

    expect(mutate()).toHaveBeenCalledTimes(2);
    const createGameCall = mutate().mock.calls[1][0];
    expect(createGameCall.variables.input.Players[0].User).toBe('RealUser');
    expect(pushMock).toHaveBeenCalledWith({ name: 'board', params: { id: 'game-1' } });
  });

  it('an empty paste leaves the submit control disabled and previewDeck uncalled', async () => {
    const wrapper = mount(QuickStartView);
    const submit = wrapper.find('[data-testid="deck-preview-submit"]');
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

    await chooseCommander(wrapper);
    await wrapper.find('[data-testid="start-table"]').trigger('click');
    await flushPromises();

    expect(mutate()).toHaveBeenCalledTimes(4);
    const guestSessionCalls = mutate().mock.calls.filter((call) => 'displayName' in (call[0].variables ?? {}));
    const createGameCalls = mutate().mock.calls.filter((call) => call[0].variables?.input?.Players);
    expect(guestSessionCalls).toHaveLength(1);
    expect(createGameCalls).toHaveLength(1);
  });

  it('restores deck text, corrections, commander picks, and display name from a stored draft', async () => {
    sessionStorage.setItem('edhgo/quickstart-draft', JSON.stringify({
      deckText: '1 Sl Ring',
      sourceURL: '',
      corrections: { 1: 'Sol Ring' },
      selectedCommanders: [{ ID: 'commander-1', Name: 'Atraxa', Text: '' }],
      displayName: 'Zoë 💫',
    }));

    const wrapper = mount(QuickStartView);
    await flushPromises();

    expect((wrapper.get('[data-testid="deck-text"]').element as HTMLTextAreaElement).value).toBe('1 Sl Ring');
    expect((wrapper.get('[data-testid="display-name"]').element as HTMLInputElement).value).toBe('Zoë 💫');
    expect(wrapper.text()).toContain('Atraxa');
    expect(wrapper.text()).toContain('Using Sol Ring');
  });

  it('a failed create preserves every field on screen and in the draft', async () => {
    const unresolvedPreview = {
      ...READY_PREVIEW,
      Unresolved: [{
        SourceLine: 1,
        RawLine: '1 Sl Ring',
        Name: 'Sl Ring',
        Reason: 'not found',
        Candidates: [{ Name: 'Sol Ring', Score: 0.99, LowConfidence: false }],
      }],
    };
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: unresolvedPreview } })
      .mockResolvedValueOnce({ data: GUEST_SESSION_RESULT })
      .mockRejectedValueOnce({
        message: 'raw SQL create failure',
        graphQLErrors: [{ message: 'raw SQL create failure', extensions: { code: 'create_error' } }],
      });

    const wrapper = mount(QuickStartView);
    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sl Ring');
    await wrapper.get('form').trigger('submit');
    await flushPromises();
    await wrapper.get('[data-testid="suggestion-chip"]').trigger('click');
    await chooseCommander(wrapper);
    await wrapper.get('[data-testid="display-name"]').setValue('Zoë 💫');
    await wrapper.get('[data-testid="start-table"]').trigger('click');
    await flushPromises();

    expect(wrapper.text()).toContain("We couldn't create your table.");
    expect(wrapper.text()).not.toContain('raw SQL create failure');
    expect((wrapper.get('[data-testid="deck-text"]').element as HTMLTextAreaElement).value).toBe('1 Sl Ring');
    expect(wrapper.text()).toContain('Atraxa');
    expect((wrapper.get('[data-testid="display-name"]').element as HTMLInputElement).value).toBe('Zoë 💫');
    expect(JSON.parse(sessionStorage.getItem('edhgo/quickstart-draft') ?? '{}')).toMatchObject({
      deckText: '1 Sl Ring',
      corrections: { 1: 'Sol Ring' },
      selectedCommanders: [{ ID: 'commander-1', Name: 'Atraxa' }],
      displayName: 'Zoë 💫',
    });
  });

  it('clears the draft before successful navigation and the next mount is empty', async () => {
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: READY_PREVIEW } })
      .mockResolvedValueOnce({ data: GUEST_SESSION_RESULT })
      .mockResolvedValueOnce({ data: CREATE_GAME_RESULT });
    const wrapper = mount(QuickStartView);
    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sol Ring');
    await wrapper.get('form').trigger('submit');
    await flushPromises();
    await chooseCommander(wrapper);
    await wrapper.get('[data-testid="start-table"]').trigger('click');
    await flushPromises();

    expect(sessionStorage.getItem('edhgo/quickstart-draft')).toBeNull();
    wrapper.unmount();

    const fresh = mount(QuickStartView);
    expect((fresh.get('[data-testid="deck-text"]').element as HTMLTextAreaElement).value).toBe('');
    expect(fresh.text()).toContain('Paste your decklist');
  });

  it('shows the static non-dismissible phone heads-up without blocking the empty funnel', () => {
    const wrapper = mount(QuickStartView);

    expect(wrapper.text()).toContain('built for tablets and larger');
    expect(wrapper.find('[data-testid="mobile-heads-up"] button').exists()).toBe(false);
    expect(wrapper.get('[data-testid="deck-preview-submit"]').attributes('disabled')).toBeDefined();
  });

  it('keeps each primary action disabled until that step is ready', async () => {
    mutate().mockResolvedValueOnce({ data: { previewDeck: READY_PREVIEW } });
    const wrapper = mount(QuickStartView);

    expect(wrapper.get('[data-testid="deck-preview-submit"]').attributes('disabled')).toBeDefined();
    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sol Ring');
    await wrapper.get('form').trigger('submit');
    await flushPromises();
    await wrapper.get('[data-testid="continue-to-commanders"]').trigger('click');
    expect(wrapper.get('[data-testid="start-table"]').attributes('disabled')).toBeDefined();
    await wrapper.get('[data-testid="commander-candidate"]').trigger('click');
    expect(wrapper.get('[data-testid="start-table"]').attributes('disabled')).toBeUndefined();
  });
});
