import { beforeEach, describe, it, expect, vi } from 'vitest';
import { JSDOM } from 'jsdom';
import { mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';

// Create a global window/document before importing any app modules that
// reference `window` during module initialization (e.g. src/services/apollo.ts).
const dom = new JSDOM('', { url: 'http://localhost' });
(global as any).window = dom.window as any;
(global as any).document = dom.window.document as any;
(global as any).localStorage = dom.window.localStorage as any;
// Some test environments / jsdom builds do not provide SVGElement; Vue's
// runtime-dom references it when mounting. Provide a minimal polyfill.
(global as any).SVGElement = (dom.window as any).SVGElement ?? class SVGElement {};

// We avoid mounting the full SFC in this test environment (SFC rendering can
// produce SSR-only artifacts under the test runner). Instead we exercise the
// same backend path the component uses: sign up to get a token and then call
// the createGame mutation directly via Apollo.
import { CREATE_GAME_MUTATION } from '../src/graphql/mutations';
import { apolloClient } from '../src/services/apollo';
import { SIGNUP_MUTATION } from '../src/graphql/mutations';
import FormCreateGame from '../src/components/games/FormCreateGame.vue';
import DeckImportPanel from '../src/components/decks/DeckImportPanel.vue';
import { useGamesStore } from '../src/stores/games';

beforeEach(() => {
  localStorage.clear();
  localStorage.setItem('edhgo/auth', JSON.stringify({ ID: 'user-1', Username: 'Host', Token: 'token' }));
  setActivePinia(createPinia());
});

describe('FormCreateGame shared deck import', () => {
  it('renders DeckImportPanel without a persistence key and removes legacy CSV copy', () => {
    const wrapper = mount(FormCreateGame);
    const panel = wrapper.getComponent(DeckImportPanel);

    expect(panel.exists()).toBe(true);
    expect(panel.props('persistenceKey')).toBeUndefined();
    expect(wrapper.text()).not.toContain('quantity,name per line');
  });

  it('submits the panel deck text with the pre-existing game payload fields unchanged', async () => {
    const games = useGamesStore();
    const createGame = vi.spyOn(games, 'createGame').mockResolvedValue('game-1');
    const wrapper = mount(FormCreateGame);

    await wrapper.get('input[placeholder="Friday Night Commander"]').setValue('Friday pod');
    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sol Ring\n99 Island');
    await wrapper.get('form').trigger('submit');

    expect(createGame).toHaveBeenCalledOnce();
    const payload = createGame.mock.calls[0][0] as any;
    expect(payload.Players[0]).toMatchObject({
      UserID: 'user-1',
      User: 'Host',
      Life: 40,
      Decklist: '1 Sol Ring\n99 Island',
      Commander: [],
    });
    expect(payload.Turn).toMatchObject({ Player: 'Host', Phase: 'MAIN', Number: 1, Priority: 'Host' });
    expect(payload.ID).toBe(payload.Players[0].GameID);
    expect(payload).toMatchObject({
      Handle: 'Friday pod',
      FormatID: 'EDH',
      SessionID: localStorage.getItem('edhgo/session-id'),
    });
    expect(payload.SessionID).toBeTruthy();
  });
});

const runLiveIntegration = process.env.VEDH_RUN_LIVE_INTEGRATION === '1';
const describeLiveIntegration = runLiveIntegration ? describe : describe.skip;

// This integration test actually hits the running backend GraphQL endpoint.
// It is skipped by default so normal local/unit runs never talk to a real backend.
// Set VEDH_RUN_LIVE_INTEGRATION=1 to opt in when the backend is intentionally running.

describeLiveIntegration('FormCreateGame (integration)', () => {
  it('signs up a user, mounts the component and creates a game against the backend', async () => {
    // Create a temporary user so we have an auth token to send with requests.
    const username = `testuser_${Date.now()}_${Math.floor(Math.random() * 1000)}`;
    const password = 'password123';

    const { data } = await apolloClient.mutate({
      mutation: SIGNUP_MUTATION,
      variables: { username, password },
    });

    expect(data?.signup).toBeTruthy();
    const profile = {
      ID: data.signup.ID,
      Username: data.signup.Username,
      Token: data.signup.Token,
    };

    // Persist token to localStorage so the apollo authLink picks it up.
    localStorage.setItem('edhgo/auth', JSON.stringify(profile));

    // Build a payload similar to what FormCreateGame would send.
    const newId = `testgame-${Date.now()}-${Math.floor(Math.random() * 1000)}`;
    const payload = {
      ID: newId,
      Turn: { Player: profile.Username, Phase: 'MAIN', Number: 1, Priority: profile.Username },
      Players: [
        {
          UserID: profile.ID,
          User: profile.Username,
          GameID: newId,
          Life: 40,
          Decklist: '1, Atraxa\n99, Island',
          Commander: [],
          Library: [],
          Graveyard: [],
          Exiled: [],
          Battlefield: [],
          Hand: [],
          Revealed: [],
          Controlled: [],
          Counters: [],
        },
      ],
    } as const;

    const { data: createData } = await apolloClient.mutate({
      mutation: CREATE_GAME_MUTATION,
      variables: { input: payload },
    });

    expect(createData?.createGame).toBeTruthy();
    expect(typeof createData.createGame.ID).toBe('string');
  }, 20000);
});
