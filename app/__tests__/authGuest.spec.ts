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

  it('claims the same identity, replaces its token once, and clears the guest credential only after success', async () => {
    mutate()
      .mockResolvedValueOnce({
        data: {
          guestSession: {
            ID: 'guest-claim-1', Username: 'Brave Sliver', DisplayName: 'Dylan', IsGuest: true,
            Token: 'guest-jwt', GuestCredential: 'guest-claim-1.secret',
          },
        },
      })
      .mockResolvedValueOnce({
        data: {
          claimGuestAccount: {
            ID: 'guest-claim-1', Username: 'dylan', DisplayName: 'Dylan', IsGuest: false, Token: 'full-jwt',
          },
        },
      });

    const auth = useAuthStore();
    await auth.createGuestSession({ sessionID: 'claim-session' });
    const claimed = await auth.claimGuestAccount({ username: ' dylan ', password: 'password-123', sessionID: 'claim-session' });

    expect(claimed).toMatchObject({ ID: 'guest-claim-1', Username: 'dylan', Token: 'full-jwt', IsGuest: false });
    expect(auth.isGuest).toBe(false);
    expect(localStorage.getItem(GUEST_CREDENTIAL_STORAGE_KEY)).toBeNull();
    expect(mutate()).toHaveBeenLastCalledWith(expect.objectContaining({
      variables: { username: 'dylan', password: 'password-123', sessionID: 'claim-session' },
    }));
  });

  it('preserves the guest profile and recovery credential when a claim conflicts', async () => {
    mutate()
      .mockResolvedValueOnce({
        data: {
          guestSession: {
            ID: 'guest-claim-2', Username: 'Swift Falcon', DisplayName: null, IsGuest: true,
            Token: 'guest-jwt-2', GuestCredential: 'guest-claim-2.secret',
          },
        },
      })
      .mockRejectedValueOnce(new Error('That username is already taken. Try another one.'));

    const auth = useAuthStore();
    await auth.createGuestSession({ sessionID: 'claim-session-2' });
    await expect(auth.claimGuestAccount({
      username: 'taken', password: 'password-123', sessionID: 'claim-session-2',
    })).rejects.toThrow('already taken');

    expect(auth.profile).toMatchObject({ ID: 'guest-claim-2', Token: 'guest-jwt-2', IsGuest: true });
    expect(localStorage.getItem(GUEST_CREDENTIAL_STORAGE_KEY)).toBe('guest-claim-2.secret');
  });

  it('rejects an identity-changing claim response without replacing the guest session', async () => {
    mutate()
      .mockResolvedValueOnce({
        data: {
          guestSession: {
            ID: 'guest-claim-3', Username: 'Mighty Griffin', DisplayName: null, IsGuest: true,
            Token: 'guest-jwt-3', GuestCredential: 'guest-claim-3.secret',
          },
        },
      })
      .mockResolvedValueOnce({
        data: {
          claimGuestAccount: {
            ID: 'different-user', Username: 'wrong', IsGuest: false, Token: 'wrong-token',
          },
        },
      });

    const auth = useAuthStore();
    await auth.createGuestSession({ sessionID: 'claim-session-3' });
    await expect(auth.claimGuestAccount({
      username: 'mighty', password: 'password-123', sessionID: 'claim-session-3',
    })).rejects.toThrow('different identity');

    expect(auth.profile).toMatchObject({ ID: 'guest-claim-3', Token: 'guest-jwt-3', IsGuest: true });
    expect(localStorage.getItem(GUEST_CREDENTIAL_STORAGE_KEY)).toBe('guest-claim-3.secret');
  });
});
