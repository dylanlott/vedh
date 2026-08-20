import { beforeEach, describe, expect, it, vi } from 'vitest';

const authStore = vi.hoisted(() => ({
  isAuthenticated: false,
  useAuthStore: vi.fn(),
}));

vi.mock('../src/stores/auth', () => ({
  useAuthStore: authStore.useAuthStore,
}));

import router from '../src/router';

beforeEach(async () => {
  authStore.isAuthenticated = false;
  authStore.useAuthStore.mockReset();
  authStore.useAuthStore.mockImplementation(() => ({
    isAuthenticated: authStore.isAuthenticated,
  }));
  await router.replace('/');
  // The landing route predates the explicit public meta marker and the
  // existing guard therefore reads auth while positioning the singleton
  // router for the next case. Each assertion starts after that setup nav.
  authStore.useAuthStore.mockClear();
});

describe('router authentication boundaries', () => {
  it.each([
    { authenticated: false, label: 'logged out' },
    { authenticated: true, label: 'logged in' },
  ])('resolves /play to quick-start without login or signup while $label', async ({ authenticated }) => {
    authStore.isAuthenticated = authenticated;

    await router.push('/play');

    expect(router.currentRoute.value.name).toBe('quick-start');
    expect(router.currentRoute.value.fullPath).toBe('/play');
    expect(router.currentRoute.value.name).not.toBe('login');
    expect(router.currentRoute.value.name).not.toBe('signup');
    expect(authStore.useAuthStore).not.toHaveBeenCalled();
  });

  it.each([
    '/games',
    '/games/game-1',
    '/games/game-1/score',
    '/games/game-1/analysis',
    '/join',
    '/join/game-1',
  ])('keeps the pre-existing requiresAuth redirect for %s', async (path) => {
    await router.push(path);

    expect(router.currentRoute.value.name).toBe('login');
    expect(router.currentRoute.value.query).toEqual({ redirect: path });
    expect(authStore.useAuthStore).toHaveBeenCalledTimes(1);
    const protectedRecord = router.getRoutes().find((record) => record.path === path.replace('game-1', ':id'));
    expect(protectedRecord?.meta.requiresAuth).toBe(true);
  });

  it.each(['/login', '/signup'])('resolves the public auth route %s without an auth-store lookup', async (path) => {
    await router.push(path);

    expect(router.currentRoute.value.path).toBe(path);
    expect(authStore.useAuthStore).not.toHaveBeenCalled();
  });
});
