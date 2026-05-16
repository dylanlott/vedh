const graphqlUrl = process.env.VEDH_GRAPHQL_URL || 'http://127.0.0.1:8080/graphql';
const timeoutMs = Number.parseInt(process.env.VEDH_SMOKE_TIMEOUT_MS || '', 10) || 10000;
const password = 'password123';
const creatorDecklist = '100, Island';
const joinerDecklist = '100, Swamp';

const SIGNUP_MUTATION = `
  mutation Signup($username: String!, $password: String!) {
    signup(username: $username, password: $password) {
      ID
      Username
      Token
    }
  }
`;

const CREATE_GAME_MUTATION = `
  mutation CreateGame($input: InputCreateGame!) {
    createGame(input: $input) {
      ID
      Players {
        ID
        Username
      }
      Turn {
        Player
        Phase
        Number
        Priority
      }
    }
  }
`;

const JOIN_GAME_MUTATION = `
  mutation JoinGame($input: InputJoinGame) {
    joinGame(input: $input) {
      ID
      Players {
        ID
        Username
      }
    }
  }
`;

function fail(message) {
  throw new Error(message);
}

function assert(condition, message) {
  if (!condition) fail(message);
}

function randomSuffix() {
  return `${Date.now()}_${Math.floor(Math.random() * 1_000_000)}`;
}

async function postGraphQL({ query, variables, token }) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(new Error(`Request timed out after ${timeoutMs}ms`)), timeoutMs);

  try {
    const response = await fetch(graphqlUrl, {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
        ...(token ? { authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify({ query, variables }),
      signal: controller.signal,
    });

    const text = await response.text();
    let payload;
    try {
      payload = text ? JSON.parse(text) : {};
    } catch (error) {
      fail(`GraphQL response was not valid JSON (${response.status}): ${text.slice(0, 200)}`);
    }

    if (!response.ok) {
      fail(`GraphQL HTTP ${response.status}: ${JSON.stringify(payload)}`);
    }

    if (payload.errors?.length) {
      const messages = payload.errors.map((entry) => entry?.message || JSON.stringify(entry)).join('; ');
      fail(`GraphQL error: ${messages}`);
    }

    return payload.data;
  } catch (error) {
    if (error?.name === 'AbortError') {
      fail(`Connection error: request to ${graphqlUrl} timed out after ${timeoutMs}ms`);
    }
    const message = error instanceof Error ? error.message : String(error);
    const causeCode = error && typeof error === 'object' && 'cause' in error && error.cause && typeof error.cause === 'object' && 'code' in error.cause
      ? error.cause.code
      : undefined;
    const causeMessage = error && typeof error === 'object' && 'cause' in error && error.cause instanceof Error
      ? error.cause.message
      : undefined;
    const details = [message, causeCode, causeMessage, graphqlUrl].filter(Boolean).join(' | ');
    fail(`Connection error: ${details}`);
  } finally {
    clearTimeout(timer);
  }
}

async function signup(username) {
  const data = await postGraphQL({
    query: SIGNUP_MUTATION,
    variables: { username, password },
  });
  assert(data?.signup?.ID, 'signup did not return ID');
  assert(data?.signup?.Username === username, 'signup returned unexpected username');
  assert(data?.signup?.Token, 'signup did not return auth token');
  return data.signup;
}

async function createGame(user, gameID) {
  const payload = {
    ID: gameID,
    Turn: {
      Player: user.Username,
      Phase: 'MAIN',
      Number: 1,
      Priority: user.Username,
    },
    Players: [
      {
        UserID: user.ID,
        User: user.Username,
        GameID: gameID,
        Life: 40,
        Decklist: creatorDecklist,
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
  };

  const data = await postGraphQL({
    query: CREATE_GAME_MUTATION,
    variables: { input: payload },
    token: user.Token,
  });

  assert(data?.createGame?.ID === gameID, 'createGame did not return expected game ID');
  assert(Array.isArray(data?.createGame?.Players), 'createGame did not return players array');
  assert(data.createGame.Players.length >= 1, 'createGame returned no players');
  return data.createGame;
}

async function joinGame(user, gameID) {
  const payload = {
    ID: gameID,
    Decklist: joinerDecklist,
    BoardState: {
      UserID: user.ID,
      User: user.Username,
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
  };

  const data = await postGraphQL({
    query: JOIN_GAME_MUTATION,
    variables: { input: payload },
    token: user.Token,
  });

  assert(data?.joinGame?.ID === gameID, 'joinGame did not return expected game ID');
  assert(Array.isArray(data?.joinGame?.Players), 'joinGame did not return players array');
  assert(data.joinGame.Players.length >= 2, `joinGame returned ${data.joinGame.Players.length} player(s), expected at least 2`);
  return data.joinGame;
}

async function main() {
  const userA = await signup(`smoke_a_${randomSuffix()}`);
  const gameID = `smoke-game-${randomSuffix()}`;
  await createGame(userA, gameID);

  const userB = await signup(`smoke_b_${randomSuffix()}`);
  const joinedGame = await joinGame(userB, gameID);

  console.log(`PASS create/join game=${joinedGame.ID} players=${joinedGame.Players.length} url=${graphqlUrl}`);
}

try {
  await main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  console.error(`FAIL ${message}`);
  process.exit(1);
}
