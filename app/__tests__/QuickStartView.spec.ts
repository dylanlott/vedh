import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';

// The Apollo client is mocked at the module boundary so this suite never
// makes a real network call, mirroring app/__tests__/productEvents.spec.ts's
// convention.
vi.mock('../src/services/apollo', () => ({
  apolloClient: { mutate: vi.fn() },
}));

const productEvents = vi.hoisted(() => ({
  getSessionID: vi.fn(() => 'browser-session-02-04'),
  track: vi.fn(),
}));
vi.mock('../src/services/productEvents', () => productEvents);

const pushMock = vi.fn();
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
}));

import { apolloClient } from '../src/services/apollo';
import DeckImportPanel from '../src/components/decks/DeckImportPanel.vue';
import { MAGIC_COMMANDER_DECK_CONTEXT } from '../src/components/decks/deckImportContext';
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

const UNRESOLVED_PREVIEW = {
  ...READY_PREVIEW,
  CardCount: 99,
  Entries: [{ Quantity: 1, Name: 'Sl Ring', Section: 'main', SourceLine: 1, Resolved: false }],
  Unresolved: [{
    SourceLine: 1,
    RawLine: '1 Sl Ring',
    Name: 'Sl Ring',
    Reason: 'not found',
    Candidates: [{ Name: 'Sol Ring', Score: 0.99, LowConfidence: false }],
  }],
};

beforeEach(() => {
  localStorage.clear();
  sessionStorage.clear();
  setActivePinia(createPinia());
  mutate().mockReset();
  pushMock.mockReset();
  productEvents.getSessionID.mockClear();
  productEvents.track.mockReset();
});

async function chooseCommander(wrapper: ReturnType<typeof mount>) {
  await wrapper.get('[data-testid="continue-to-commanders"]').trigger('click');
  await wrapper.get('[data-testid="commander-candidate"]').trigger('click');
}

function operationCalls(kind: 'guest' | 'create') {
  return mutate().mock.calls.filter((call) => {
    const variables = call[0].variables ?? {};
    if (kind === 'guest') return 'displayName' in variables && 'sessionID' in variables;
    return Boolean(variables.input?.Players);
  });
}

function expectPreservedActivationState(wrapper: ReturnType<typeof mount>, rawMessage: string) {
  expect((wrapper.get('[data-testid="deck-text"]').element as HTMLTextAreaElement).value).toBe('1 Sl Ring');
  expect(wrapper.text()).toMatch(/(?:Sl Ring→Sol Ring|Line 1: Using Sol Ring)/);
  expect(wrapper.text()).toContain('Atraxa');
  expect((wrapper.get('[data-testid="display-name"]').element as HTMLInputElement).value).toBe('Zoë 💫');
  expect(wrapper.text()).not.toContain(rawMessage);
  expect(JSON.parse(sessionStorage.getItem('edhgo/quickstart-draft') ?? '{}')).toMatchObject({
    deckText: '1 Sl Ring',
    corrections: { 1: 'Sol Ring' },
    selectedCommanders: [{ ID: 'commander-1', Name: 'Atraxa' }],
    displayName: 'Zoë 💫',
  });
}

describe('QuickStartView', () => {
  it('emits quick_start_viewed exactly once per mount and never emits an authoritative name', () => {
    const first = mount(QuickStartView);
    expect(productEvents.track).toHaveBeenCalledTimes(1);
    expect(productEvents.track).toHaveBeenLastCalledWith('quick_start_viewed');
    first.unmount();

    const second = mount(QuickStartView);
    expect(productEvents.track).toHaveBeenCalledTimes(2);
    expect(productEvents.track.mock.calls.map(([name]) => name)).toEqual([
      'quick_start_viewed',
      'quick_start_viewed',
    ]);
    expect(productEvents.track.mock.calls.flat()).not.toEqual(expect.arrayContaining([
      'game_created',
      'guest_session_created',
    ]));
    second.unmount();
  });

  it('renders the paste surface for a logged-out visitor with no navigation to login/signup', () => {
    const wrapper = mount(QuickStartView);
    const panel = wrapper.getComponent(DeckImportPanel);

    expect(wrapper.text()).toContain('Paste your decklist');
    expect(panel.props('context')).toEqual(MAGIC_COMMANDER_DECK_CONTEXT);
    const context = wrapper.get('[data-testid="deck-import-context"]');
    expect(context.findAll('dt').map((term) => term.text())).toEqual(['Game', 'Format']);
    expect(context.findAll('dd').map((value) => value.text())).toEqual([
      'Magic: The Gathering',
      'Commander (EDH)',
    ]);
    expect(wrapper.get('[data-testid="deck-import-panel"]').attributes('aria-label')).toBe(
      'Magic: The Gathering Commander (EDH) deck import',
    );
    expect(pushMock).not.toHaveBeenCalled();
  });

  it('previewDeck is called once with { text, sessionID } and the totals line renders', async () => {
    mutate().mockResolvedValueOnce({ data: { previewDeck: READY_PREVIEW } });

    const wrapper = mount(QuickStartView);
    await wrapper.find('textarea').setValue('1, Sol Ring');
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
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
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();

    await chooseCommander(wrapper);
    await wrapper.find('[data-testid="start-table"]').trigger('click');
    await flushPromises();

    expect(mutate()).toHaveBeenCalledTimes(3);
    const guestCall = mutate().mock.calls[1][0];
    expect(guestCall.variables.sessionID).toBeTruthy();
    const createGameCall = mutate().mock.calls[2][0];
    expect(createGameCall.variables.input.Players[0].User).toBe('Brave Sliver');
    expect(createGameCall.variables.input).toMatchObject({
      Handle: 'Commander table',
      FormatID: 'EDH',
      SessionID: 'browser-session-02-04',
    });
    expect(createGameCall.variables.input.Players[0].Decklist).toBe('1, Sol Ring');

    expect(pushMock).toHaveBeenCalledWith({ name: 'board', params: { id: 'game-1' } });
  });

  it('emits exactly the three permitted client events in order across a complete run', async () => {
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: READY_PREVIEW } })
      .mockResolvedValueOnce({ data: GUEST_SESSION_RESULT })
      .mockResolvedValueOnce({ data: CREATE_GAME_RESULT });

    const wrapper = mount(QuickStartView);
    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sol Ring');
    expect(productEvents.track.mock.calls.map(([name]) => name)).toEqual(['quick_start_viewed']);

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    await chooseCommander(wrapper);
    await wrapper.get('[data-testid="start-table"]').trigger('click');
    await flushPromises();

    const names = productEvents.track.mock.calls.map(([name]) => name);
    expect(names).toEqual(['quick_start_viewed', 'deck_import_started', 'game_create_started']);
    expect(new Set(names)).toEqual(new Set(['quick_start_viewed', 'deck_import_started', 'game_create_started']));
    expect(productEvents.track.mock.invocationCallOrder[2]).toBeLessThan(mutate().mock.invocationCallOrder[2]);
  });

  it('with an authenticated auth store: the same path runs and guestSession is never called', async () => {
    const auth = useAuthStore();
    auth.profile = { ID: 'real-user-1', Username: 'RealUser', Token: 'real-jwt' };

    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: READY_PREVIEW } })
      .mockResolvedValueOnce({ data: CREATE_GAME_RESULT });

    const wrapper = mount(QuickStartView);
    await wrapper.find('textarea').setValue('1, Sol Ring');
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();

    await chooseCommander(wrapper);
    await wrapper.find('[data-testid="start-table"]').trigger('click');
    await flushPromises();

    expect(mutate()).toHaveBeenCalledTimes(2);
    const createGameCall = mutate().mock.calls[1][0];
    expect(createGameCall.variables.input.Players[0].User).toBe('RealUser');
    expect(operationCalls('guest')).toHaveLength(0);
    expect(pushMock).toHaveBeenCalledWith({ name: 'board', params: { id: 'game-1' } });
  });

  it('does not create a guest on mount, paste, or preview and creates one only at the final valid step', async () => {
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: READY_PREVIEW } })
      .mockResolvedValueOnce({ data: GUEST_SESSION_RESULT })
      .mockResolvedValueOnce({ data: CREATE_GAME_RESULT });
    const wrapper = mount(QuickStartView);

    expect(operationCalls('guest')).toHaveLength(0);
    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sol Ring');
    expect(operationCalls('guest')).toHaveLength(0);
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    expect(operationCalls('guest')).toHaveLength(0);

    await chooseCommander(wrapper);
    await wrapper.get('[data-testid="start-table"]').trigger('click');
    await flushPromises();
    expect(operationCalls('guest')).toHaveLength(1);
  });

  it('submits the panel canonical deck after replacement without mutating adjacent pasted lines', async () => {
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: UNRESOLVED_PREVIEW } })
      .mockResolvedValueOnce({ data: GUEST_SESSION_RESULT })
      .mockResolvedValueOnce({ data: CREATE_GAME_RESULT });
    const wrapper = mount(QuickStartView);

    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sl Ring\n1 Island');
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    await wrapper.get('[data-testid="suggestion-control"]').setValue(true);
    await chooseCommander(wrapper);
    await wrapper.get('[data-testid="start-table"]').trigger('click');
    await flushPromises();

    const createCall = operationCalls('create')[0][0];
    expect(createCall.variables.input.Players[0].Decklist).toBe('1 Sol Ring\n1 Island');
    expect(createCall.variables.input.Players[0].Decklist).not.toContain('Sl Ring');
  });

  it('an empty paste leaves the submit control disabled and previewDeck uncalled', async () => {
    const wrapper = mount(QuickStartView);
    const submit = wrapper.find('[data-testid="deck-preview-submit"]');
    expect(submit.attributes('disabled')).toBeDefined();

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
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
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    // Submitting the identical paste again (e.g. a second click before
    // noticing the first request already resolved) re-runs previewDeck but
    // must not multiply anything downstream.
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();

    await chooseCommander(wrapper);
    await wrapper.find('[data-testid="start-table"]').trigger('click');
    await flushPromises();

    expect(mutate()).toHaveBeenCalledTimes(4);
    const guestSessionCalls = mutate().mock.calls.filter((call) => 'displayName' in (call[0].variables ?? {}));
    const createGameCalls = mutate().mock.calls.filter((call) => call[0].variables?.input?.Players);
    expect(guestSessionCalls).toHaveLength(1);
    expect(createGameCalls).toHaveLength(1);
    expect(productEvents.track.mock.calls.map(([name]) => name)).toEqual([
      'quick_start_viewed',
      'deck_import_started',
      'deck_import_started',
      'game_create_started',
    ]);
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
    expect(wrapper.text()).toContain('Line 1: Using Sol Ring');

    mutate().mockResolvedValueOnce({ data: { previewDeck: UNRESOLVED_PREVIEW } });
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    expect(wrapper.get('[data-testid="suggestion-control"]').attributes('checked')).toBeDefined();
    expect(wrapper.get('[data-testid="correction-confirmation"]').text()).toContain('Sl Ring→Sol Ring');
  });

  it('a failed create preserves every field on screen and in the draft', async () => {
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: UNRESOLVED_PREVIEW } })
      .mockResolvedValueOnce({ data: GUEST_SESSION_RESULT })
      .mockRejectedValueOnce({
        message: 'raw SQL create failure',
        graphQLErrors: [{ message: 'raw SQL create failure', extensions: { code: 'create_error' } }],
      });

    const wrapper = mount(QuickStartView);
    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sl Ring');
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    await wrapper.get('[data-testid="suggestion-control"]').setValue(true);
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

  it.each([
    {
      code: 'provider_unavailable',
      title: "We can't reach that deck link right now.",
      affordance: 'Paste instead',
      raw: 'raw provider response 503 secret',
    },
    {
      code: 'preview_error',
      title: "We couldn't fully read a few cards.",
      affordance: 'Fix these cards',
      raw: 'raw GraphQL preview resolver trace',
    },
  ])('preserves the full draft and safe recovery for $code', async ({ code, title, affordance, raw }) => {
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: UNRESOLVED_PREVIEW } })
      .mockRejectedValueOnce({
        message: raw,
        graphQLErrors: [{ message: raw, extensions: { code } }],
      });

    const wrapper = mount(QuickStartView);
    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sl Ring');
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    await wrapper.get('[data-testid="suggestion-control"]').setValue(true);
    await chooseCommander(wrapper);
    await wrapper.get('[data-testid="display-name"]').setValue('Zoë 💫');

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();

    expect(wrapper.text()).toContain(title);
    expect(wrapper.text()).toContain(affordance);
    expectPreservedActivationState(wrapper, raw);
  });

  it.each([
    {
      stage: 'guest session',
      code: 'guest_session_error',
      title: "We couldn't set up your seat at the table.",
      raw: 'raw guest SQL duplicate-key detail',
    },
    {
      stage: 'game create',
      code: 'create_error',
      title: "We couldn't create your table.",
      raw: 'raw create SQL statement and bind values',
    },
  ])('preserves the full draft and safe recovery for a $stage failure', async ({ code, title, raw }) => {
    mutate().mockResolvedValueOnce({ data: { previewDeck: UNRESOLVED_PREVIEW } });
    if (code === 'create_error') mutate().mockResolvedValueOnce({ data: GUEST_SESSION_RESULT });
    mutate().mockRejectedValueOnce({
      message: raw,
      graphQLErrors: [{ message: raw, extensions: { code } }],
    });

    const wrapper = mount(QuickStartView);
    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sl Ring');
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    await wrapper.get('[data-testid="suggestion-control"]').setValue(true);
    await chooseCommander(wrapper);
    await wrapper.get('[data-testid="display-name"]').setValue('Zoë 💫');
    await wrapper.get('[data-testid="start-table"]').trigger('click');
    await flushPromises();

    expect(wrapper.text()).toContain(title);
    expect(wrapper.text()).toContain('Try again');
    expectPreservedActivationState(wrapper, raw);
  });

  it('retries create with the established guest profile instead of minting another guest', async () => {
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: READY_PREVIEW } })
      .mockResolvedValueOnce({ data: GUEST_SESSION_RESULT })
      .mockRejectedValueOnce({ graphQLErrors: [{ extensions: { code: 'create_error' } }] })
      .mockResolvedValueOnce({ data: CREATE_GAME_RESULT });

    const wrapper = mount(QuickStartView);
    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sol Ring');
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    await chooseCommander(wrapper);
    await wrapper.get('[data-testid="start-table"]').trigger('click');
    await flushPromises();
    await wrapper.get('[data-testid="quick-start-error"] button').trigger('click');
    await flushPromises();

    expect(operationCalls('guest')).toHaveLength(1);
    expect(operationCalls('create')).toHaveLength(2);
    expect(pushMock).toHaveBeenCalledWith({ name: 'board', params: { id: 'game-1' } });
  });

  it('disables Start my table while createGame is in flight and ignores a double press', async () => {
    let resolveCreate!: (value: typeof CREATE_GAME_RESULT extends infer T ? { data: T } : never) => void;
    const pendingCreate = new Promise<{ data: typeof CREATE_GAME_RESULT }>((resolve) => {
      resolveCreate = resolve;
    });
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: READY_PREVIEW } })
      .mockResolvedValueOnce({ data: GUEST_SESSION_RESULT })
      .mockReturnValueOnce(pendingCreate);

    const wrapper = mount(QuickStartView);
    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sol Ring');
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    await chooseCommander(wrapper);

    const start = wrapper.get('[data-testid="start-table"]');
    await start.trigger('click');
    await flushPromises();
    expect(start.attributes('disabled')).toBeDefined();
    await start.trigger('click');
    expect(operationCalls('create')).toHaveLength(1);

    resolveCreate({ data: CREATE_GAME_RESULT });
    await flushPromises();
    expect(operationCalls('guest')).toHaveLength(1);
    expect(operationCalls('create')).toHaveLength(1);
  });

  it('round-trips astral-plane and combining characters through the restore path unchanged', async () => {
    const deckText = '1 Yoshimaru, Ever Faithful 𠜎\n99 I\u0301sland';
    const displayName = 'A\u030Asa 𐐷';
    sessionStorage.setItem('edhgo/quickstart-draft', JSON.stringify({
      deckText,
      sourceURL: '',
      corrections: {},
      selectedCommanders: [],
      displayName,
    }));

    const wrapper = mount(QuickStartView);
    await flushPromises();
    expect((wrapper.get('[data-testid="deck-text"]').element as HTMLTextAreaElement).value).toBe(deckText);
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    expect(JSON.parse(sessionStorage.getItem('edhgo/quickstart-draft') ?? '{}')).toMatchObject({ deckText, displayName });
  });

  it('clears the draft before successful navigation and the next mount is empty', async () => {
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: READY_PREVIEW } })
      .mockResolvedValueOnce({ data: GUEST_SESSION_RESULT })
      .mockResolvedValueOnce({ data: CREATE_GAME_RESULT });
    const wrapper = mount(QuickStartView);
    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sol Ring');
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
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
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    await wrapper.get('[data-testid="continue-to-commanders"]').trigger('click');
    expect(wrapper.get('[data-testid="start-table"]').attributes('disabled')).toBeDefined();
    await wrapper.get('[data-testid="commander-candidate"]').trigger('click');
    expect(wrapper.get('[data-testid="start-table"]').attributes('disabled')).toBeUndefined();
  });
});
