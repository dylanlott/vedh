<template>
  <section class="auth-card">
    <h1>Log in</h1>
    <form @submit.prevent="handleSubmit">
      <label>
        <span>Username</span>
        <input v-model="form.username" autocomplete="username" required />
      </label>
      <label>
        <span>Password</span>
        <input v-model="form.password" type="password" autocomplete="current-password" required />
      </label>
      <button class="primary" :disabled="auth.loading">{{ auth.loading ? 'Signing in…' : 'Log in' }}</button>
      <p class="helper" v-if="auth.errorMessage">{{ auth.errorMessage }}</p>
    </form>
    <p class="switcher">
      Need an account?
      <RouterLink to="/signup">Create one</RouterLink>
    </p>
  </section>
</template>

<script setup lang="ts">
import { reactive } from 'vue';
import { useRouter, RouterLink, useRoute } from 'vue-router';
import { useAuthStore } from '../stores/auth';

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

const form = reactive({
  username: '',
  password: '',
});

async function handleSubmit() {
  if (!form.username || !form.password) return;
  try {
    await auth.login({ username: form.username, password: form.password });
    const redirect = (route.query.redirect as string) ?? '/games';
    router.push(redirect);
  } catch (error) {
    // Error handled in store
  }
}
</script>

<style scoped lang="scss">
.auth-card {
  max-width: 420px;
  margin: 3rem auto;
  padding: 2.5rem;
  border-radius: 20px;
  background: rgba(15, 19, 26, 0.85);
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 24px 48px -24px rgba(0, 0, 0, 0.45);
}

h1 {
  margin-top: 0;
  margin-bottom: 1.5rem;
}

form {
  display: grid;
  gap: 1.25rem;
}

label {
  display: grid;
  gap: 0.5rem;
}

input {
  padding: 0.75rem 1rem;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.06);
  color: inherit;
}

button.primary {
  border: none;
  border-radius: 10px;
  padding: 0.75rem 1rem;
  font-size: 1rem;
  font-weight: 600;
  background: linear-gradient(120deg, #85d7ff, #3f8cff);
  color: #0b1016;
  cursor: pointer;
  transition: filter 0.2s ease;
}

button.primary:disabled {
  filter: grayscale(0.5);
  cursor: progress;
}

.helper {
  color: #ff9d9d;
  margin: 0;
}

.switcher {
  margin-top: 1.5rem;
}

.switcher a {
  color: #85d7ff;
}
</style>
