import { beforeEach, describe, expect, it, vi } from 'vitest';
import { flushPromises, mount } from '@vue/test-utils';

vi.mock('../src/services/apollo', () => ({
  apolloClient: { query: vi.fn() },
}));

import { apolloClient } from '../src/services/apollo';
import CommanderReview from '../src/components/decks/CommanderReview.vue';
import { partnerConstraintMessage } from '../src/services/commanderPartner';

type QueryMock = ReturnType<typeof vi.fn>;

const solo = { ID: 'solo', Name: 'Atraxa, Praetors’ Voice', Text: 'Flying, vigilance' };
const pir = {
  ID: 'pir',
  Name: 'Pir, Imaginative Rascal',
  Text: 'Partner with Toothy, Imaginary Friend (When this creature enters...)',
};
const toothy = {
  ID: 'toothy',
  Name: 'Toothy, Imaginary Friend',
  Text: 'Partner with Pir, Imaginative Rascal (When this creature enters...)',
};
const wrongPartner = { ID: 'wrong', Name: 'Wrong Partner', Text: 'Partner' };

beforeEach(() => {
  (apolloClient.query as unknown as QueryMock).mockReset();
});

describe('CommanderReview', () => {
  it('renders the contract empty state and keeps manual search available', () => {
    const wrapper = mount(CommanderReview, { props: { candidates: [], selected: [] } });

    expect(wrapper.text()).toContain('No legal commander detected yet');
    expect(wrapper.text()).toContain('None of your parsed cards can legally be a Commander');
    expect(wrapper.find('[data-testid="commander-search"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="candidate-list-spinner"]').exists()).toBe(false);
    expect(wrapper.find('[data-testid="candidate-list-error"]').exists()).toBe(false);
  });

  it('renders up to three candidates and marks only the chosen card selected', async () => {
    const wrapper = mount(CommanderReview, {
      props: { candidates: [solo, pir, toothy, wrongPartner], selected: [] },
    });

    expect(wrapper.findAll('[data-testid="commander-candidate"]')).toHaveLength(3);
    await wrapper.findAll('[data-testid="commander-candidate"]')[0].trigger('click');

    expect(wrapper.findAll('.selected')).toHaveLength(1);
    expect(wrapper.findAll('[aria-pressed="true"]')).toHaveLength(1);
    expect(wrapper.emitted('selection-change')?.at(-1)?.[0]).toEqual([solo]);
    expect(wrapper.attributes('data-valid-selection')).toBe('true');
  });

  it('uses the single-commander path for one candidate with no partner prompt', () => {
    const wrapper = mount(CommanderReview, { props: { candidates: [solo], selected: [] } });

    expect(wrapper.findAll('[data-testid="commander-candidate"]')).toHaveLength(1);
    expect(wrapper.find('[data-testid="partner-constraint"]').exists()).toBe(false);
  });

  it('rejects an illegal second commander with the shared helper copy', async () => {
    const wrapper = mount(CommanderReview, {
      props: { candidates: [pir, wrongPartner], selected: [] },
    });

    await wrapper.findAll('[data-testid="commander-candidate"]')[0].trigger('click');
    await wrapper.findAll('[data-testid="commander-candidate"]')[1].trigger('click');

    expect(wrapper.get('[data-testid="partner-constraint"]').text()).toBe(partnerConstraintMessage(pir));
    expect(wrapper.attributes('data-valid-selection')).toBe('false');
    expect(wrapper.emitted('selection-change')?.at(-1)?.[0]).toEqual([pir]);
  });

  it('accepts a legal partner pair and never permits a third selection', async () => {
    const wrapper = mount(CommanderReview, {
      props: { candidates: [pir, toothy, wrongPartner], selected: [] },
    });

    await wrapper.findAll('[data-testid="commander-candidate"]')[0].trigger('click');
    await wrapper.findAll('[data-testid="commander-candidate"]')[1].trigger('click');
    await wrapper.findAll('[data-testid="commander-candidate"]')[2].trigger('click');

    expect(wrapper.emitted('selection-change')?.at(-1)?.[0]).toEqual([pir, toothy]);
    expect(wrapper.attributes('data-valid-selection')).toBe('true');
  });

  it('searches all cards with the existing typeahead shape', async () => {
    vi.useFakeTimers();
    (apolloClient.query as unknown as QueryMock).mockResolvedValueOnce({ data: { search: [solo] } });
    const wrapper = mount(CommanderReview, { props: { candidates: [], selected: [] } });

    await wrapper.get('[data-testid="commander-search"]').setValue('Atraxa');
    await vi.advanceTimersByTimeAsync(150);
    await flushPromises();

    expect(apolloClient.query).toHaveBeenCalledOnce();
    expect(wrapper.findAll('[data-testid="manual-commander-result"]')).toHaveLength(1);
    vi.useRealTimers();
  });

  it('truncates long names with a matching title and exposes capped overflow classes', () => {
    const longName = `💫A\u0301 ${'Legendary Commander '.repeat(5)}`;
    const wrapper = mount(CommanderReview, {
      props: { candidates: [{ ID: 'long', Name: longName, Text: '' }], selected: [] },
    });

    const label = wrapper.get('[data-testid="candidate-name"]');
    expect(label.attributes('title')).toBe(longName);
    expect(label.text()).not.toContain('\uFFFD');
    expect(wrapper.get('[data-testid="commander-candidates"]').classes()).toContain('typeahead');
  });
});
