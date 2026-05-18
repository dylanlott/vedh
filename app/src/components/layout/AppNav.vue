<template>
  <header class="app-nav">
    <div class="brand" @click="goHome">
      <span class="logo">vEDH</span>
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
  background: rgba(var(--vedh-bg-rgb), 0.86);
  backdrop-filter: blur(12px);
  color: var(--vedh-text);
  z-index: 20;
  border-bottom: 1px solid var(--vedh-border);
}

.brand {
  display: flex;
  align-items: baseline;
  font-weight: 700;
  cursor: pointer;
}

.logo {
  font-size: 1.25rem;
  letter-spacing: 0.08em;
  color: var(--vedh-primary);
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
  position: relative;
  transition: color 0.2s ease;
}

a::after {
  content: '';
  position: absolute;
  left: 0;
  bottom: -0.3rem;
  width: 100%;
  height: 2px;
  border-radius: 999px;
  background: var(--vedh-primary-gradient);
  opacity: 0;
  transform: scaleX(0.6);
  transform-origin: left;
  transition: opacity 0.2s ease, transform 0.2s ease;
}

a:hover,
a:focus-visible {
  color: var(--vedh-primary);
}

a:hover::after,
a:focus-visible::after {
  opacity: 1;
  transform: scaleX(1);
}

a.router-link-active {
  color: var(--vedh-primary);
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
  background: var(--vedh-primary-gradient);
  color: var(--vedh-primary-contrast);
}

button.secondary {
  background: rgba(255, 244, 237, 0.1);
  color: var(--vedh-text);
  border: 1px solid var(--vedh-border);
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
  background: rgba(255, 244, 237, 0.08);
  padding: 0.35rem 0.75rem;
  border-radius: 999px;
  border: 1px solid var(--vedh-border);
}
</style>
