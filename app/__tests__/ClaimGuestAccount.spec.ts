import { beforeEach, describe, expect, it, vi } from 'vitest';
import { flushPromises, mount } from '@vue/test-utils';

const auth = vi.hoisted(() => ({
  profile: { ID: 'guest-1', Username: 'Brave Sliver', DisplayName: 'Table Mage', IsGuest: true, Token: 'guest-token' },
  claimGuestAccount: vi.fn(),
}));
vi.mock('../src/stores/auth', () => ({ useAuthStore: () => auth }));

const productEvents = vi.hoisted(() => ({ getSessionID: vi.fn(() => 'session-1'), track: vi.fn() }));
vi.mock('../src/services/productEvents', () => productEvents);

import ClaimGuestAccount from '../src/components/auth/ClaimGuestAccount.vue';

beforeEach(() => {
  sessionStorage.clear();
  auth.claimGuestAccount.mockReset();
  productEvents.track.mockReset();
});

describe('ClaimGuestAccount', () => {
  it('prefills the display name, validates locally, and reports a retryable conflict in place', async () => {
    auth.claimGuestAccount.mockRejectedValueOnce(new Error('That username is already taken. Try another one.'));
    const wrapper = mount(ClaimGuestAccount, { props: { gameID: 'game-1', role: 'host' } });

    expect(wrapper.get('input[name="username"]').element.value).toBe('Table Mage');
    await wrapper.get('form').trigger('submit');
    expect(wrapper.get('[role="alert"]').text()).toBe('Choose a username and password.');
    expect(auth.claimGuestAccount).not.toHaveBeenCalled();

    await wrapper.get('input[name="password"]').setValue('password-123');
    await wrapper.get('form').trigger('submit');
    await flushPromises();
    expect(wrapper.get('[role="alert"]').text()).toContain('already taken');
    expect(wrapper.text()).toContain('Save your games');
    expect(productEvents.track).toHaveBeenCalledWith('account_claim_started');
  });

  it('keeps the board context and renders a quiet success after claiming', async () => {
    auth.claimGuestAccount.mockResolvedValueOnce({ ID: 'guest-1', Username: 'tablemage', IsGuest: false, Token: 'full-token' });
    const wrapper = mount(ClaimGuestAccount, { props: { gameID: 'game-1', role: 'invitee' } });
    await wrapper.get('input[name="username"]').setValue('tablemage');
    await wrapper.get('input[name="password"]').setValue('password-123');
    await wrapper.get('form').trigger('submit');
    await flushPromises();

    expect(auth.claimGuestAccount).toHaveBeenCalledWith({
      username: 'tablemage', password: 'password-123', sessionID: 'session-1',
    });
    expect(wrapper.get('[role="status"]').text()).toContain('keep playing');
  });

  it('remembers dismissal only for the current game and never calls the claim', async () => {
    const wrapper = mount(ClaimGuestAccount, { props: { gameID: 'game-dismissed', role: 'host' } });
    await wrapper.get('[aria-label="Dismiss save games prompt"]').trigger('click');
    expect(wrapper.find('form').exists()).toBe(false);
    expect(sessionStorage.getItem('edhgo/claim-dismissed/game-dismissed')).toBe('true');
    expect(auth.claimGuestAccount).not.toHaveBeenCalled();
  });
});
