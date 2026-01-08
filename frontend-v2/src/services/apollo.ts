import { ApolloClient, createHttpLink, InMemoryCache, split } from '@apollo/client/core';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { getMainDefinition } from '@apollo/client/utilities';
import { createClient } from 'graphql-ws';
import { setContext } from '@apollo/client/link/context';
import type { DocumentNode } from 'graphql';

// Prefer Vite dev proxy during development to avoid cross-origin issues.
let httpUri = import.meta.env.VITE_GRAPHQL_HTTP;
let wsUri = import.meta.env.VITE_GRAPHQL_WS;

if (import.meta.env.DEV) {
  if (!httpUri) httpUri = '/graphql';
  if (!wsUri) {
    // Guard access to `window` because tests may run in a node environment
    // before a DOM global exists. Only attempt to build a wsUri from the
    // current location if `window` is available.
    if (typeof window !== 'undefined' && window.location) {
      const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
      wsUri = `${proto}://${window.location.host}/graphql`;
    }
  }
}
// Fallbacks for non-dev environments
if (!httpUri) httpUri = 'http://localhost:8080/graphql';
if (!wsUri) wsUri = httpUri.replace(/^http/, 'ws');

// If we're running in a Node/test environment (no window) and the httpUri is
// a relative path like '/graphql', make it absolute so the Node fetch
// implementation can parse it.
if (typeof window === 'undefined' && httpUri && httpUri.startsWith('/')) {
  httpUri = `http://localhost:8080${httpUri}`;
  wsUri = httpUri.replace(/^http/, 'ws');
}

const httpLink = createHttpLink({
  uri: httpUri,
  // We authenticate via Authorization header, not cookies.
  // Avoid CORS errors with wildcard origins on the server.
  credentials: 'omit',
});

const authLink = setContext((_operation, { headers }) => {
  // Pinia store is not yet available here during initial import, so we read
  // directly from localStorage. In test environments `localStorage` may not
  // exist on the global, so guard access and fall back to window.localStorage
  // when available.
  const getRawAuth = () => {
    try {
      if (typeof localStorage !== 'undefined' && localStorage) return localStorage.getItem('edhgo/auth');
    } catch (e) {
      // ignore
    }
    try {
      if (typeof window !== 'undefined' && window.localStorage) return window.localStorage.getItem('edhgo/auth');
    } catch (e) {
      // ignore
    }
    return null;
  };

  const raw = getRawAuth();
  let token: string | undefined;
  if (raw) {
    try {
      const parsed = JSON.parse(raw) as { Token?: string };
      token = parsed.Token;
    } catch (error) {
      console.warn('[apollo] failed to parse auth token', error);
    }
  }

  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : undefined,
    },
  };
});

// Create a WebSocket link only when we have a proper wsUri. In some test
// environments there is no window location or wsUri and creating a ws link
// (which may access global networking) is undesirable.
let wsLink: any = null;
// graphql-ws requires a WebSocket implementation in Node.
// For vitest/unit tests, we don’t need subscriptions, so skip wsLink.
const hasWebSocket = typeof WebSocket !== 'undefined';
if (wsUri && hasWebSocket) {
  wsLink = new GraphQLWsLink(
    createClient({
      url: wsUri,
      connectionParams: () => {
        const raw = ((): string | null => {
          try {
            if (typeof localStorage !== 'undefined' && localStorage) return localStorage.getItem('edhgo/auth');
          } catch (e) {}
          try {
            if (typeof window !== 'undefined' && window.localStorage) return window.localStorage.getItem('edhgo/auth');
          } catch (e) {}
          return null;
        })();
        if (!raw) return {};
        try {
          const parsed = JSON.parse(raw) as { Token?: string };
          return parsed.Token ? { authorization: `Bearer ${parsed.Token}` } : {};
        } catch (error) {
          console.warn('[apollo] failed to parse auth token for ws connection', error);
          return {};
        }
      },
    }),
  );
}

const link = wsLink
  ? split(
      ({ query }: { query: DocumentNode }) => {
        const definition = getMainDefinition(query);
        return definition.kind === 'OperationDefinition' && definition.operation === 'subscription';
      },
      wsLink,
      authLink.concat(httpLink),
    )
  : authLink.concat(httpLink);

export const apolloClient = new ApolloClient({
  link,
  cache: new InMemoryCache(),
  devtools: { enabled: import.meta.env.DEV },
});
