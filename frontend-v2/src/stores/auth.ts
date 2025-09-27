import { defineStore } from 'pinia';
import { ref, computed, watch } from 'vue';
import { apolloClient } from '../services/apollo';
import { LOGIN_MUTATION, SIGNUP_MUTATION } from '../graphql/mutations';
import type { LoginMutation, LoginMutationVariables, SignupMutation, SignupMutationVariables } from '../types/generated';

interface AuthProfile {
  ID: string;
  Username: string;
  Token: string;
}

const STORAGE_KEY = 'edhgo/auth';

function loadPersistedProfile(): AuthProfile | null {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as AuthProfile;
  } catch (error) {
    console.warn('[auth] failed to parse profile:', error);
    localStorage.removeItem(STORAGE_KEY);
    return null;
  }
}

export const useAuthStore = defineStore('auth', () => {
  const profile = ref<AuthProfile | null>(loadPersistedProfile());
  const loading = ref(false);
  const errorMessage = ref<string | null>(null);

  watch(profile, (value: AuthProfile | null) => {
    if (!value) {
      localStorage.removeItem(STORAGE_KEY);
      return;
    }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(value));
  }, { deep: true });

  const isAuthenticated = computed(() => Boolean(profile.value?.Token));

  async function login(credentials: { username: string; password: string; redirect?: string }) {
    loading.value = true;
    errorMessage.value = null;
    try {
      const { data } = await apolloClient.mutate<LoginMutation, LoginMutationVariables>({
        mutation: LOGIN_MUTATION,
        variables: {
          username: credentials.username,
          password: credentials.password,
        },
      });
      if (!data?.login) {
        throw new Error('Login returned empty response');
      }
      profile.value = {
        ID: data.login.ID,
        Username: data.login.Username,
        Token: data.login.Token,
      };
      return profile.value;
    } catch (error: unknown) {
      console.error('[auth] login failed', error);
      errorMessage.value = error instanceof Error ? error.message : 'Login failed';
      throw error;
    } finally {
      loading.value = false;
    }
  }

  async function signup(payload: { username: string; password: string }) {
    loading.value = true;
    errorMessage.value = null;
    try {
      const { data } = await apolloClient.mutate<SignupMutation, SignupMutationVariables>({
        mutation: SIGNUP_MUTATION,
        variables: payload,
      });
      if (!data?.signup) {
        throw new Error('Signup returned empty response');
      }
      profile.value = {
        ID: data.signup.ID,
        Username: data.signup.Username,
        Token: data.signup.Token,
      };
      return profile.value;
    } catch (error: unknown) {
      console.error('[auth] signup failed', error);
      errorMessage.value = error instanceof Error ? error.message : 'Signup failed';
      throw error;
    } finally {
      loading.value = false;
    }
  }

  function logout() {
    profile.value = null;
  }

  return {
    profile,
    loading,
    errorMessage,
    isAuthenticated,
    login,
    signup,
    logout,
  };
});
