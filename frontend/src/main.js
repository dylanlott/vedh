import Vue from 'vue'
import VueApollo from 'vue-apollo'
import VueCookies from 'vue-cookies'
// import VueMatomo from 'vue-matomo'
import Buefy from 'buefy'
import 'buefy/dist/buefy.css'

import api from '@/gqlclient'
import App from './App.vue'
import router from './router'
import { store } from './store'

Vue.config.productionTip = false
Vue.use(Buefy)

Vue.use(VueCookies)
Vue.$cookies.config('30d', null, null, null, 'Strict')

const apolloProvider = new VueApollo({
  defaultClient: api,
});
Vue.use(VueApollo)

const vm = new Vue({
  router,
  store,
  provide: apolloProvider.provide(),
  render: (h) => h(App),
});
vm.$mount('#app');
