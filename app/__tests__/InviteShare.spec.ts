import { beforeEach, describe, expect, it, vi } from 'vitest';
import { flushPromises, mount } from '@vue/test-utils';

const productEvents = vi.hoisted(() => ({ track: vi.fn() }));
vi.mock('../src/services/productEvents', () => productEvents);

import InviteShare from '../src/components/InviteShare.vue';

function setNavigatorShare(value: unknown) {
  Object.defineProperty(navigator, 'share', { configurable: true, value });
}

function setClipboard(value: unknown) {
  Object.defineProperty(navigator, 'clipboard', { configurable: true, value });
}

beforeEach(() => {
  productEvents.track.mockReset();
  setNavigatorShare(undefined);
  setClipboard(undefined);
});

describe('InviteShare', () => {
  it('prefers native share and records only the method plus dedicated event fields', async () => {
    const share = vi.fn().mockResolvedValue(undefined);
    const writeText = vi.fn();
    setNavigatorShare(share);
    setClipboard({ writeText });
    const wrapper = mount(InviteShare, { props: { gameID: 'game-1', role: 'host' } });

    await wrapper.get('[data-testid="share-invite"]').trigger('click');
    await flushPromises();

    expect(share).toHaveBeenCalledWith({ title: 'Join my vEDH table', url: 'http://localhost:3000/join/game-1' });
    expect(writeText).not.toHaveBeenCalled();
    expect(productEvents.track).toHaveBeenCalledWith('invite_copied', { share_method: 'native_share' }, {
      gameID: 'game-1', role: 'host', source: 'board',
    });
    expect(JSON.stringify(productEvents.track.mock.calls)).not.toContain('/join/game-1');
    expect(wrapper.text()).toContain('Invite shared.');
  });

  it('falls back from native share to clipboard copy', async () => {
    setNavigatorShare(vi.fn().mockRejectedValue(new Error('cancelled')));
    const writeText = vi.fn().mockResolvedValue(undefined);
    setClipboard({ writeText });
    const wrapper = mount(InviteShare, { props: { gameID: 'game-2', role: 'invitee' } });

    await wrapper.get('[data-testid="share-invite"]').trigger('click');
    await flushPromises();

    expect(writeText).toHaveBeenCalledWith('http://localhost:3000/join/game-2');
    expect(productEvents.track).toHaveBeenCalledWith('invite_copied', { share_method: 'clipboard' }, {
      gameID: 'game-2', role: 'invitee', source: 'board',
    });
    expect(wrapper.text()).toContain('Invite link copied.');
  });

  it('exposes a selectable manual URL when automatic sharing fails', async () => {
    setNavigatorShare(vi.fn().mockRejectedValue(new Error('blocked')));
    setClipboard({ writeText: vi.fn().mockRejectedValue(new Error('blocked')) });
    const wrapper = mount(InviteShare, { props: { gameID: 'game-3', role: 'host' } });

    await wrapper.get('[data-testid="share-invite"]').trigger('click');
    await flushPromises();

    expect(wrapper.text()).toContain('Automatic sharing was blocked.');
    expect((wrapper.get('.manual-invite input').element as HTMLInputElement).value).toBe('http://localhost:3000/join/game-3');
    expect(productEvents.track).not.toHaveBeenCalled();
  });
});
