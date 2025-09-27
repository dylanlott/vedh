<template>
  <header class="app-nav">
    <div class="brand" @click="goHome">
      <span class="logo">EDH</span>
      <span class="wordmark">Go</span>
      <span class="badge">v2</span>
    </div>
    <nav>
      <RouterLink to="/">Home</RouterLink>
      <RouterLink to="/games" v-if="isAuthenticated">Games</RouterLink>
      <RouterLink to="/join" v-if="isAuthenticated">Join Game</RouterLink>
    </nav>
    <div class="auth">
      <button v-if="!isAuthenticated" class="secondary" @click="goLogin">Log in</button>
      <button v-if="!isAuthenticated" class="primary" @click="goSignup">Sign up</button>
      <div v-else class="user-chip">
        <span>{{ username }}</span>
        <button class="link" @click="logout">Log out</button>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRouter, RouterLink } from 'vue-router';
import { useAuthStore } from '../../stores/auth';

const auth = useAuthStore();
const router = useRouter();

const isAuthenticated = computed(() => auth.isAuthenticated);
const username = computed(() => auth.profile?.Username ?? '');

function goHome() {
  router.push('/');
}

function goLogin() {
  router.push('/login');
}

function goSignup() {
  router.push('/signup');
}

function logout() {
  auth.logout();
  router.push('/');
}
</script>

<style scoped lang="scss">
.app-nav {
  position: sticky;
  top: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1.5rem;
  background: rgba(10, 10, 10, 0.85);
  backdrop-filter: blur(12px);
  color: #f5f5f5;
  z-index: 20;
}

.brand {
  display: flex;
  align-items: baseline;
  gap: 0.35rem;
  font-weight: 700;
  cursor: pointer;
}

.logo {
  font-size: 1.25rem;
  letter-spacing: 0.08em;
}

.wordmark {
  font-size: 1.25rem;
  color: #85d7ff;
}

.badge {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: rgba(255, 255, 255, 0.6);
}

nav {
  display: flex;
  gap: 1rem;
  align-items: center;
}

a {
  color: inherit;
  text-decoration: none;
  font-weight: 500;
}

a.router-link-active {
  color: #85d7ff;
}

.auth {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

button {
  cursor: pointer;
  border: none;
  border-radius: 6px;
  padding: 0.4rem 0.9rem;
  font-weight: 600;
  transition: transform 0.15s ease;
}

button:hover {
  transform: translateY(-1px);
}

button.primary {
  background: #85d7ff;
  color: #111;
}

button.secondary {
  background: rgba(255, 255, 255, 0.12);
  color: #f5f5f5;
}

button.link {
  background: none;
  color: inherit;
  padding: 0;
}

.user-chip {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: rgba(255, 255, 255, 0.08);
  padding: 0.35rem 0.75rem;
  border-radius: 999px;
}
</style>
