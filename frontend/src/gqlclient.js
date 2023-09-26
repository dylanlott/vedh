import fetch from 'unfetch';
import { ApolloClient } from 'apollo-client';
import { HttpLink } from 'apollo-link-http';
import { WebSocketLink } from 'apollo-link-ws';
import { split } from 'apollo-link';
import { getMainDefinition } from 'apollo-utilities';
import { InMemoryCache } from 'apollo-cache-inmemory';
const cache = new InMemoryCache({
    addTypename: false
})
console.log('base url: ', process.env.VUE_APP_BASE_URL)

const httpLink = new HttpLink({
    uri: process.env.VUE_APP_BASE_URL,
    fetch: fetch
});
console.log('websocket url: ', process.env.VUE_APP_WEBSOCKET_URL)
const wsLink = new WebSocketLink({
    uri: process.env.VUE_APP_WEBSOCKET_URL,
    options: {
        reconnect: true,
    },
});
const link = split(
    ({ query }) => {
        const { kind, operation } = getMainDefinition(query);
        return kind === 'OperationDefinition' && operation === 'subscription';
    },
    wsLink,
    httpLink,
);
export default new ApolloClient({
    link: link,
    cache: cache,
});
