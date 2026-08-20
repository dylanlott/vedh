import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { nextTick } from 'vue';

// The Apollo client is mocked at the module boundary so this suite never
// makes a real network call, mirroring app/__tests__/productEvents.spec.ts's
// convention.
vi.mock('../src/services/apollo', () => ({
  apolloClient: { mutate: vi.fn() },
}));

import { apolloClient } from '../src/services/apollo';
import { useAuthStore } from '../src/stores/auth';
import { getSessionID } from '../src/services/productEvents';

const AUTH_STORAGE_KEY = 'edhgo/auth';
const GUEST_CREDENTIAL_STORAGE_KEY = 'edhgo/guest-credential';
const SESSION_ID_STORAGE_KEY = 'edhgo/session-id';

type MutateMock = ReturnType<typeof vi.fn>;

function mutate(): MutateMock {
  return apolloClient.mutate as unknown as MutateMock;
}

beforeEach(() => {
  localStorage.clear();
  setActivePinia(createPinia());
  mutate().mockReset();
});

describe('guest credential storage', () => {
  it('writes the credential under its own key and keeps it out of the persisted auth profile', async () => {
    mutate().mockResolvedValueOnce({
      data: {
        guestSession: {
          ID: 'guest-1',
          Username: 'Brave Sliver',
          DisplayName: null,
          IsGuest: true,
          Token: 'guest-jwt',
          GuestCredential: 'guest-1.super-secret',
        },
      },
    });

    const auth = useAuthStore();
    await auth.createGuestSession({ sessionID: 'session-abc' });

    expect(localStorage.getItem(GUEST_CREDENTIAL_STORAGE_KEY)).toBe('guest-1.super-secret');

    const persistedProfile = JSON.parse(localStorage.getItem(AUTH_STORAGE_KEY) ?? '{}');
    expect(persistedProfile).not.toHaveProperty('GuestCredential');
    expect(persistedProfile.Token).toBe('guest-jwt');
  });

  it('leaves the analytics session identifier untouched and distinct from the guest credential', async () => {
    const sessionIDBefore = getSessionID();

    mutate().mockResolvedValueOnce({
      data: {
        guestSession: {
          ID: 'guest-2',
          Username: 'Mighty Griffin',
          DisplayName: null,
          IsGuest: true,
          Token: 'guest-jwt-2',
          GuestCredential: 'guest-2.other-secret',
        },
      },
    });

    const auth = useAuthStore();
    await auth.createGuestSession({ sessionID: sessionIDBefore });

    const sessionIDAfter = localStorage.getItem(SESSION_ID_STORAGE_KEY);
    expect(sessionIDAfter).toBe(sessionIDBefore);

    const guestCredential = localStorage.getItem(GUEST_CREDENTIAL_STORAGE_KEY);
    expect(guestCredential).not.toBe(sessionIDAfter);
  });

  it('readGuestCredential returns null when nothing is stored and the stored value when present', () => {
    const auth = useAuthStore();
    expect(auth.readGuestCredential()).toBeNull();

    localStorage.setItem(GUEST_CREDENTIAL_STORAGE_KEY, 'a-stored-credential');
    expect(auth.readGuestCredential()).toBe('a-stored-credential');
  });

  it('readGuestCredential degrades to null when the storage accessor throws', () => {
    const auth = useAuthStore();
    const getItemSpy = vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('storage disabled');
    });
    try {
      expect(auth.readGuestCredential()).toBeNull();
    } finally {
      getItemSpy.mockRestore();
    }
  });

  it('logout clears both the auth profile and the guest credential', async () => {
    mutate().mockResolvedValueOnce({
      data: {
        guestSession: {
          ID: 'guest-3',
          Username: 'Swift Falcon',
          DisplayName: null,
          IsGuest: true,
          Token: 'guest-jwt-3',
          GuestCredential: 'guest-3.secret',
        },
      },
    });

    const auth = useAuthStore();
    await auth.createGuestSession({ sessionID: 'session-xyz' });
    expect(localStorage.getItem(GUEST_CREDENTIAL_STORAGE_KEY)).toBeTruthy();

    auth.logout();
    // profile is watched with { deep: true } (the default flush timing),
    // so the localStorage removal side effect runs on the next tick.
    await nextTick();

    expect(auth.profile).toBeNull();
    expect(localStorage.getItem(AUTH_STORAGE_KEY)).toBeNull();
    expect(localStorage.getItem(GUEST_CREDENTIAL_STORAGE_KEY)).toBeNull();
  });
});
