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
    const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
    wsUri = `${proto}://${window.location.host}/graphql`;
  }
}
// Fallbacks for non-dev environments
if (!httpUri) httpUri = 'http://localhost:8080/graphql';
if (!wsUri) wsUri = httpUri.replace(/^http/, 'ws');

const httpLink = createHttpLink({
  uri: httpUri,
  // We authenticate via Authorization header, not cookies.
  // Avoid CORS errors with wildcard origins on the server.
  credentials: 'omit',
});

const authLink = setContext((_operation, { headers }) => {
  // Pinia store is not yet available here during initial import, so we read directly.
  const raw = localStorage.getItem('edhgo/auth');
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

const wsLink = new GraphQLWsLink(createClient({
  url: wsUri,
  connectionParams: () => {
    const raw = localStorage.getItem('edhgo/auth');
    if (!raw) return {};
    try {
      const parsed = JSON.parse(raw) as { Token?: string };
      return parsed.Token ? { authorization: `Bearer ${parsed.Token}` } : {};
    } catch (error) {
      console.warn('[apollo] failed to parse auth token for ws connection', error);
      return {};
    }
  },
}));

const link = split(
  ({ query }: { query: DocumentNode }) => {
    const definition = getMainDefinition(query);
    return definition.kind === 'OperationDefinition' && definition.operation === 'subscription';
  },
  wsLink,
  authLink.concat(httpLink),
);

export const apolloClient = new ApolloClient({
  link,
  cache: new InMemoryCache(),
  devtools: { enabled: import.meta.env.DEV },
});
