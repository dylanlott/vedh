import { createRouter, createWebHistory, RouteRecordRaw, type NavigationGuardNext, type RouteLocationNormalized } from 'vue-router';
import { useAuthStore } from '../stores/auth';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'landing',
    component: () => import('../views/LandingView.vue'),
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/LoginView.vue'),
    meta: { public: true },
  },
  {
    path: '/signup',
    name: 'signup',
    component: () => import('../views/SignupView.vue'),
    meta: { public: true },
  },
  {
    path: '/games',
    name: 'games',
    component: () => import('../views/GamesView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/games/:id/analysis',
    name: 'game-analysis',
    component: () => import('../views/GameAnalysisView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/games/:id',
    name: 'board',
    component: () => import('../views/BoardView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/games/:id/score',
    name: 'score',
    component: () => import('../views/ScoreView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/card/:id',
    name: 'card',
    component: () => import('../views/CardView.vue'),
  },
  {
    path: '/join',
    name: 'join',
    component: () => import('../views/JoinGameView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/join/:id',
    name: 'join-game',
    component: () => import('../views/JoinGameView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/games/404',
    name: 'game-not-found',
    component: () => import('../views/GameDoesNotExistView.vue'),
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/games/404',
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) => {
  if (to.meta.public) {
    return next();
  }
  const auth = useAuthStore();
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return next({ name: 'login', query: { redirect: to.fullPath } });
  }
  return next();
});

export default router;
