import Vue from 'vue';
import Router from 'vue-router';

// router and store access for authentication
import router from '@/router.js';
import { Users } from '@/store'

// router views
import Landing from '@/components/Landing.vue';
import Login from '@/components/Login.vue';
import Signup from '@/components/Signup.vue';
import Games from '@/components/Games.vue';
import GameDoesNotExist from '@/components/GameDoesNotExist.vue';
import Board from '@/components/Board.vue';
import Card from '@/components/Card.vue';
import Score from '@/components/Score.vue';
import JoinGame from '@/components/JoinGame.vue';

Vue.use(Router);

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: 'home',
      component: Landing,
    },
    {
      path: '/login',
      name: 'login',
      component: Login,
    },
    {
      path: '/signup',
      name: 'signup',
      component: Signup
    },
    {
      path: '/games',
      name: 'games',
      component: Games,
      meta: {
        auth: true,
      },
    },
    {
      path: '/games/404',
      name: 'GameDoesNotExist',
      component: GameDoesNotExist
    },
    {
      path: '/games/:id',
      name: 'board',
      component: Board,
    },
    {
      path: '/games/:id/score',
      name: 'score_screen',
      component: Score,
    },
    {
      path: '/card/:id',
      name: 'card',
      component: Card,
    },
    {
      path: '/join/:id',
      name: 'join',
      component: JoinGame,
    }
  ],
});

// handle auth before each route change
router.beforeEach((to, from, next) => {
  if ((to.matched.length > 0) && (to.matched[0].meta.auth)) {
    const authed = Users.getters.authenticated
    if (!authed) {
      return next({
        path: '/login',
        query: {
          redirect: to.fullPath
        }
      })
    }
    return next()
  }
  return next()
})