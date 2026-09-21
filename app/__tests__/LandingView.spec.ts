import { beforeEach, describe, expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';

const router = vi.hoisted(() => ({ push: vi.fn() }));
vi.mock('vue-router', () => ({ useRouter: () => router }));

const productEvents = vi.hoisted(() => ({ track: vi.fn() }));
vi.mock('../src/services/productEvents', () => productEvents);

import LandingView from '../src/views/LandingView.vue';

beforeEach(() => {
  router.push.mockReset();
  productEvents.track.mockReset();
});

describe('LandingView', () => {
  it('states the Commander activation path without unsupported claims', () => {
    const wrapper = mount(LandingView);
    expect(wrapper.get('h1').text()).toBe('Paste a Commander deck. Start a table.');
    expect(wrapper.text()).toContain('Guest-first hosting');
    expect(wrapper.text()).toContain('Open an invite');
    expect(wrapper.text()).not.toContain('any TCG');
  });

  it('records allowlisted attribution on the primary CTA and routes directly to quick start', async () => {
    const wrapper = mount(LandingView);
    await wrapper.findAll('button').find(button => button.text() === 'Start a table')!.trigger('click');
    expect(productEvents.track).toHaveBeenCalledWith('landing_primary_cta', {}, { source: 'landing' });
    expect(router.push).toHaveBeenCalledWith({ name: 'quick-start' });
  });

  it('keeps login and invite paths available without emitting the primary event', async () => {
    const wrapper = mount(LandingView);
    await wrapper.findAll('button').find(button => button.text() === 'Log in')!.trigger('click');
    await wrapper.findAll('button').find(button => button.text() === 'Open an invite')!.trigger('click');
    expect(router.push).toHaveBeenNthCalledWith(1, { name: 'login' });
    expect(router.push).toHaveBeenNthCalledWith(2, { name: 'join' });
    expect(productEvents.track).not.toHaveBeenCalled();
  });
});
