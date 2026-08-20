import { beforeEach, describe, it, expect, vi } from 'vitest';
import { JSDOM } from 'jsdom';
import { mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';

const pushMock = vi.fn();
vi.mock('vue-router', () => ({
  useRoute: () => ({ params: { id: 'game-1' } }),
  useRouter: () => ({ push: pushMock }),
}));

// Setup minimal DOM and storage before importing modules that use window/localStorage
const dom = new JSDOM('', { url: 'http://localhost' });
(global as any).window = dom.window as any;
(global as any).document = dom.window.document as any;
(global as any).localStorage = dom.window.localStorage as any;
(global as any).SVGElement = (dom.window as any).SVGElement ?? class SVGElement {};

import { apolloClient } from '../src/services/apollo';
import { SIGNUP_MUTATION, CREATE_GAME_MUTATION, JOIN_GAME_MUTATION } from '../src/graphql/mutations';
import JoinGameView from '../src/views/JoinGameView.vue';
import DeckImportPanel from '../src/components/decks/DeckImportPanel.vue';
import { useGamesStore } from '../src/stores/games';

beforeEach(() => {
  localStorage.clear();
  localStorage.setItem('edhgo/auth', JSON.stringify({ ID: 'user-2', Username: 'Joiner', Token: 'token' }));
  setActivePinia(createPinia());
  pushMock.mockReset();
});

describe('JoinGame shared deck import', () => {
  it('renders DeckImportPanel without a persistence key and removes legacy CSV copy', () => {
    const wrapper = mount(JoinGameView);
    const panel = wrapper.getComponent(DeckImportPanel);

    expect(panel.exists()).toBe(true);
    expect(panel.props('persistenceKey')).toBeUndefined();
    expect(wrapper.text()).not.toContain('quantity,name per line');
  });

  it('submits the same panel deck representation through the unchanged join payload', async () => {
    const games = useGamesStore();
    const joinGame = vi.spyOn(games, 'joinGame').mockResolvedValue('game-1');
    const wrapper = mount(JoinGameView);

    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sol Ring\n99 Island');
    await wrapper.get('form').trigger('submit');

    expect(joinGame).toHaveBeenCalledOnce();
    expect(joinGame.mock.calls[0][0]).toMatchObject({
      ID: 'game-1',
      Decklist: '1 Sol Ring\n99 Island',
      BoardState: {
        UserID: 'user-2',
        User: 'Joiner',
        GameID: 'game-1',
        Life: 40,
        Commander: [],
      },
    });
    expect(pushMock).toHaveBeenCalledWith({ name: 'board', params: { id: 'game-1' } });
  });
});

const runLiveIntegration = process.env.VEDH_RUN_LIVE_INTEGRATION === '1';
const describeLiveIntegration = runLiveIntegration ? describe : describe.skip;

// This integration test talks to the live GraphQL backend.
// It is skipped by default so routine test runs stay local/safe.
// Set VEDH_RUN_LIVE_INTEGRATION=1 to opt in when you explicitly want live coverage.

describeLiveIntegration('JoinGame (integration)', () => {
  it('creates a game as user A and joins it as user B', async () => {
    // Sign up user A
    const usernameA = `userA_${Date.now()}_${Math.floor(Math.random() * 1000)}`;
    const password = 'password123';

    const { data: signupA } = await apolloClient.mutate({
      mutation: SIGNUP_MUTATION,
      variables: { username: usernameA, password },
    });

    expect(signupA?.signup).toBeTruthy();
    const aProfile = {
      ID: signupA!.signup.ID,
      Username: signupA!.signup.Username,
      Token: signupA!.signup.Token,
    } as const;

    // Use user A token for game creation
    localStorage.setItem('edhgo/auth', JSON.stringify(aProfile));

    const gameID = `join-test-${Date.now()}-${Math.floor(Math.random() * 1000)}`;
    const createPayload = {
      ID: gameID,
      Turn: { Player: aProfile.Username, Phase: 'MAIN', Number: 1, Priority: aProfile.Username },
      Players: [
        {
          UserID: aProfile.ID,
          User: aProfile.Username,
          GameID: gameID,
          Life: 40,
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

    const { data: created } = await apolloClient.mutate({
      mutation: CREATE_GAME_MUTATION,
      variables: { input: createPayload },
    });

    expect(created?.createGame?.ID).toBe(gameID);

    // Sign up user B
    const usernameB = `userB_${Date.now()}_${Math.floor(Math.random() * 1000)}`;
    const { data: signupB } = await apolloClient.mutate({
      mutation: SIGNUP_MUTATION,
      variables: { username: usernameB, password },
    });
    expect(signupB?.signup).toBeTruthy();
    const bProfile = {
      ID: signupB!.signup.ID,
      Username: signupB!.signup.Username,
      Token: signupB!.signup.Token,
    } as const;

    // Switch auth to user B
    localStorage.setItem('edhgo/auth', JSON.stringify(bProfile));

    const joinPayload = {
      ID: gameID,
      Decklist: '',
      BoardState: {
        UserID: bProfile.ID,
        User: bProfile.Username,
        GameID: gameID,
        Life: 40,
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
    } as const;

    const { data: joined } = await apolloClient.mutate({
      mutation: JOIN_GAME_MUTATION,
      variables: { input: joinPayload },
    });

    expect(joined?.joinGame?.ID).toBe(gameID);
    // Expect the returned game to contain at least two players now
    expect(joined?.joinGame?.Players?.length).toBeGreaterThanOrEqual(2);
  }, 25000);
});
