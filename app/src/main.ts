import { createApp, h } from 'vue';
import { createPinia } from 'pinia';
import { DefaultApolloClient } from '@vue/apollo-composable';

import App from './App.vue';
import router from './router';
import { apolloClient } from './services/apollo';
import './styles/main.scss';

const app = createApp({
  setup() {
    return () => h(App);
  },
});

app.use(createPinia());
app.use(router);
app.provide(DefaultApolloClient, apolloClient);
app.mount('#app');
