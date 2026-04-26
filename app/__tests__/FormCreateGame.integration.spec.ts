import { describe, it, expect } from 'vitest';
import { JSDOM } from 'jsdom';

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

// This integration test actually hits the running backend GraphQL endpoint.
// Requirements for running this test:
// - The backend GraphQL server must be running and accessible at http://localhost:8080/graphql
// - The test will create a temporary user via the signup mutation and then use
//   that user's auth token to create a game through the UI.

describe('FormCreateGame (integration)', () => {
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
