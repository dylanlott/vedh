import Vue from 'vue'
import Vuex from 'vuex'
import VueApollo from 'vue-apollo'
import api from '@/gqlclient'
import Buefy from 'buefy'
import 'buefy/dist/buefy.css'
import './scss/custom.scss'
import VueCookies from 'vue-cookies'
import router from './router'
import App from './App.vue'
import { Cards, Boardstates, Games, Users } from './store'
import { AuthPlugin } from './auth'
import VueMatomo from 'vue-matomo'

Vue.use(Buefy)
Vue.use(VueCookies)
Vue.$cookies.config('30d', null, null, null, 'Strict')
Vue.config.productionTip = false
const apolloProvider = new VueApollo({
  defaultClient: api,
});
Vue.use(VueApollo)
Vue.use(AuthPlugin)
Vue.use(VueMatomo, {
  router: router,
  host: 'https://analytics.edhgo.com/',
  siteId: 1,
  requireConsent: true,
  requireCookieConsent: true,
  enableHeartBeatTimer: true,
})
var store = new Vuex.Store({
  modules: {
    Boardstates,
    Cards,
    Games,
    Users,
  }
})
const vm = new Vue({
  router,
  store,
  provide: apolloProvider.provide(),
  render: (h) => h(App),
});
vm.$mount('#app');
