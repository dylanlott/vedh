import { describe, it, expect } from 'vitest';
import { JSDOM } from 'jsdom';

// Setup minimal DOM and storage before importing modules that use window/localStorage
const dom = new JSDOM('', { url: 'http://localhost' });
(global as any).window = dom.window as any;
(global as any).document = dom.window.document as any;
(global as any).localStorage = dom.window.localStorage as any;
(global as any).SVGElement = (dom.window as any).SVGElement ?? class SVGElement {};

import { apolloClient } from '../src/services/apollo';
import { SIGNUP_MUTATION, CREATE_GAME_MUTATION, JOIN_GAME_MUTATION } from '../src/graphql/mutations';

describe('JoinGame (integration)', () => {
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
      Turn: { Player: aProfile.Username, Phase: 'MAIN', Number: 1 },
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
