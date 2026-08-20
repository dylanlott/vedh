import { defineStore } from 'pinia';
import { ref, computed, watch } from 'vue';
import { apolloClient } from '../services/apollo';
import { LOGIN_MUTATION, SIGNUP_MUTATION, GUEST_SESSION_MUTATION } from '../graphql/mutations';
import type {
  LoginMutation,
  LoginMutationVariables,
  SignupMutation,
  SignupMutationVariables,
  GuestSessionMutation,
  GuestSessionMutationVariables,
} from '../types/generated';

interface AuthProfile {
  ID: string;
  Username: string;
  Token: string;
  DisplayName?: string;
  IsGuest?: boolean;
}

const STORAGE_KEY = 'edhgo/auth';
const GUEST_CREDENTIAL_STORAGE_KEY = 'edhgo/guest-credential';

function getStorage(): Storage | null {
  try {
    if (typeof localStorage !== 'undefined' && localStorage) return localStorage;
  } catch {
    // Ignore and try the window-scoped accessor below.
  }
  try {
    if (typeof window !== 'undefined' && window.localStorage) return window.localStorage;
  } catch {
    // Storage is unavailable in this environment.
  }
  return null;
}

export function readGuestCredential(): string | null {
  try {
    return getStorage()?.getItem(GUEST_CREDENTIAL_STORAGE_KEY) ?? null;
  } catch {
    return null;
  }
}

function writeGuestCredential(value: string): void {
  try {
    getStorage()?.setItem(GUEST_CREDENTIAL_STORAGE_KEY, value);
  } catch {
    // A disabled storage surface must not prevent the guest session itself.
  }
}

function clearGuestCredential(): void {
  try {
    getStorage()?.removeItem(GUEST_CREDENTIAL_STORAGE_KEY);
  } catch {
    // A disabled storage surface is already effectively cleared.
  }
}

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

  async function createGuestSession(args: { displayName?: string; sessionID: string }) {
    loading.value = true;
    errorMessage.value = null;
    try {
      const { data } = await apolloClient.mutate<GuestSessionMutation, GuestSessionMutationVariables>({
        mutation: GUEST_SESSION_MUTATION,
        variables: {
          displayName: args.displayName,
          sessionID: args.sessionID,
        },
      });
      if (!data?.guestSession) {
        throw new Error('GuestSession returned empty response');
      }
      if (!data.guestSession.GuestCredential) {
        throw new Error('GuestSession returned empty credential');
      }
      writeGuestCredential(data.guestSession.GuestCredential);
      // GuestCredential is intentionally omitted here: it must never ride
      // inside the persisted auth profile (D-2.4).
      profile.value = {
        ID: data.guestSession.ID,
        Username: data.guestSession.Username,
        Token: data.guestSession.Token,
        DisplayName: data.guestSession.DisplayName ?? undefined,
        IsGuest: data.guestSession.IsGuest ?? undefined,
      };
      return profile.value;
    } catch (error: unknown) {
      console.error('[auth] guest session failed', error);
      errorMessage.value = error instanceof Error ? error.message : 'Guest session failed';
      throw error;
    } finally {
      loading.value = false;
    }
  }

  function logout() {
    profile.value = null;
    clearGuestCredential();
  }

  return {
    profile,
    loading,
    errorMessage,
    isAuthenticated,
    login,
    signup,
    createGuestSession,
    readGuestCredential,
    logout,
  };
});
