import { beforeEach, describe, expect, it, vi } from 'vitest';
import { flushPromises, mount } from '@vue/test-utils';

vi.mock('../src/services/apollo', () => ({
  apolloClient: { mutate: vi.fn(), query: vi.fn() },
}));

import { apolloClient } from '../src/services/apollo';
import DeckImportPanel from '../src/components/decks/DeckImportPanel.vue';

type Mock = ReturnType<typeof vi.fn>;

function mutate(): Mock {
  return apolloClient.mutate as unknown as Mock;
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
}

function preview(overrides: Record<string, unknown> = {}) {
  return {
    SourceType: 'PASTE',
    CardCount: 1,
    Entries: [{ Quantity: 1, Name: 'Sol Ring', Section: 'main', SourceLine: 1, Resolved: true }],
    CommanderCandidates: [],
    Unresolved: [],
    Warnings: [],
    CanContinue: true,
    BlockingErrors: [],
    ...overrides,
  };
}

function apolloError(code: string, raw: string) {
  return { message: raw, graphQLErrors: [{ message: raw, extensions: { code } }] };
}

beforeEach(() => {
  mutate().mockReset();
  (apolloClient.query as unknown as Mock).mockReset();
});

describe('DeckImportPanel', () => {
  it('renders its empty-state copy, disables submit, and never previews empty input', async () => {
    const wrapper = mount(DeckImportPanel, { props: { sessionId: 'session-1' } });

    expect(wrapper.text()).toContain('Paste your decklist');
    expect(wrapper.text()).toContain('Paste a list from any format you already use');
    expect(wrapper.get('[data-testid="deck-preview-submit"]').attributes('disabled')).toBeDefined();

    await wrapper.get('form').trigger('submit');
    expect(mutate()).not.toHaveBeenCalled();
    expect(wrapper.find('[data-testid="preview-totals"]').exists()).toBe(false);
  });

  it('renders one panel-level loading state and disables submit while previewDeck is pending', async () => {
    const pending = deferred<{ data: { previewDeck: ReturnType<typeof preview> } }>();
    mutate().mockReturnValueOnce(pending.promise);
    const wrapper = mount(DeckImportPanel, { props: { sessionId: 'session-1' } });

    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sol Ring');
    await wrapper.get('form').trigger('submit');

    expect(wrapper.text()).toContain('Checking your deck');
    expect(wrapper.find('[data-testid="preview-spinner"]').exists()).toBe(true);
    expect(wrapper.get('[data-testid="deck-preview-submit"]').attributes('disabled')).toBeDefined();

    pending.resolve({ data: { previewDeck: preview() } });
    await flushPromises();
    expect(wrapper.find('[data-testid="preview-spinner"]').exists()).toBe(false);
  });

  it.each([
    [0, 'No warnings'],
    [1, '1 warning'],
    [3, '3 warnings'],
  ])('pluralizes %i warnings in a resolved totals card', async (count, copy) => {
    mutate().mockResolvedValueOnce({ data: { previewDeck: preview({ Warnings: Array(count).fill('warning') }) } });
    const wrapper = mount(DeckImportPanel, { props: { sessionId: 'session-1', initialText: '1 Sol Ring' } });

    expect(wrapper.find('[data-testid="preview-totals"]').exists()).toBe(false);
    await wrapper.get('form').trigger('submit');
    await flushPromises();

    expect(wrapper.get('[data-testid="preview-totals"]').text()).toContain('1 of 1 card ready');
    expect(wrapper.get('[data-testid="preview-totals"]').text()).toContain(copy);
  });

  it.each([
    [0, 'No blocking errors'],
    [1, '1 blocking error'],
    [3, '3 blocking errors'],
  ])('pluralizes %i blocking errors and gates Continue', async (count, copy) => {
    mutate().mockResolvedValueOnce({
      data: { previewDeck: preview({ BlockingErrors: Array(count).fill('blocking'), CanContinue: true }) },
    });
    const wrapper = mount(DeckImportPanel, { props: { sessionId: 'session-1', initialText: '1 Sol Ring' } });

    await wrapper.get('form').trigger('submit');
    await flushPromises();

    expect(wrapper.get('[data-testid="preview-totals"]').text()).toContain(copy);
    expect(wrapper.get('[data-testid="continue-to-commanders"]').attributes('disabled')).toBe(
      count === 0 ? undefined : '',
    );
  });

  it('hides zero unresolved rows and renders one or many inside a capped list', async () => {
    const issue = (line: number) => ({
      SourceLine: line,
      RawLine: `${line} Typo Card`,
      Name: 'Typo Card',
      Reason: 'not found',
      Candidates: [
        { Name: 'Nearest Card', Score: 0.9, LowConfidence: false },
        { Name: 'Second Card', Score: 0.8, LowConfidence: false },
        { Name: 'Third Card', Score: 0.7, LowConfidence: true },
        { Name: 'Fourth Card', Score: 0.6, LowConfidence: true },
      ],
    });
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: preview() } })
      .mockResolvedValueOnce({ data: { previewDeck: preview({ Unresolved: [issue(1)] }) } })
      .mockResolvedValueOnce({ data: { previewDeck: preview({ Unresolved: [issue(1), issue(2), issue(3)] }) } });
    const wrapper = mount(DeckImportPanel, { props: { sessionId: 'session-1', initialText: '1 Typo Card' } });

    await wrapper.get('form').trigger('submit');
    await flushPromises();
    expect(wrapper.find('[data-testid="unresolved-list"]').exists()).toBe(false);

    await wrapper.get('form').trigger('submit');
    await flushPromises();
    expect(wrapper.text()).toContain('1 card needs a quick look');
    expect(wrapper.findAll('[data-testid="suggestion-chip"]')).toHaveLength(3);
    expect(wrapper.find('[data-testid="manual-card-name"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="remove-unresolved-line"]').exists()).toBe(true);

    await wrapper.get('form').trigger('submit');
    await flushPromises();
    expect(wrapper.findAll('[data-testid="unresolved-row"]')).toHaveLength(3);
    expect(wrapper.get('[data-testid="unresolved-list"]').classes()).toContain('typeahead');
  });

  it('applies a suggestion only on click and preserves the correction across a re-preview', async () => {
    const unresolved = [{
      SourceLine: 1,
      RawLine: '1 Sl Ring',
      Name: 'Sl Ring',
      Reason: 'not found',
      Candidates: [{ Name: 'Sol Ring', Score: 0.98, LowConfidence: false }],
    }];
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: preview({ Unresolved: unresolved }) } })
      .mockResolvedValueOnce({ data: { previewDeck: preview({ Unresolved: unresolved }) } });
    const wrapper = mount(DeckImportPanel, { props: { sessionId: 'session-1', initialText: '1 Sl Ring' } });

    await wrapper.get('form').trigger('submit');
    await flushPromises();
    expect(wrapper.emitted('change')?.at(-1)?.[0]).toMatchObject({ corrections: {} });

    await wrapper.get('[data-testid="suggestion-chip"]').trigger('click');
    expect(wrapper.emitted('change')?.at(-1)?.[0]).toMatchObject({ corrections: { 1: 'Sol Ring' } });

    await wrapper.get('form').trigger('submit');
    await flushPromises();
    expect(wrapper.emitted('change')?.at(-1)?.[0]).toMatchObject({ corrections: { 1: 'Sol Ring' } });
    expect(mutate().mock.calls[1][0].variables.input.text).toBe('1 Sl Ring');
  });

  it.each([
    ['provider_unavailable', 'Paste instead'],
    ['preview_error', 'Fix these cards'],
  ])('renders safe %s recovery without row errors or totals', async (code, affordance) => {
    const raw = `GraphQL SQL provider internals for ${code}`;
    mutate().mockRejectedValueOnce(apolloError(code, raw));
    const wrapper = mount(DeckImportPanel, { props: { sessionId: 'session-1', initialText: '1 Sol Ring' } });

    await wrapper.get('form').trigger('submit');
    await flushPromises();

    expect(wrapper.findAll('[data-testid="activation-error"]')).toHaveLength(1);
    expect(wrapper.findAll('[data-testid="row-error"]')).toHaveLength(0);
    expect(wrapper.find('[data-testid="preview-totals"]').exists()).toBe(false);
    expect(wrapper.text()).toContain(affordance);
    expect(wrapper.text()).not.toContain(raw);
  });

  it('preserves pasted Unicode byte-for-byte and truncates names on code-point boundaries', async () => {
    const pasted = '1 A\u0301ether 💫 Adept';
    mutate().mockResolvedValueOnce({
      data: { previewDeck: preview({ Unresolved: [{
        SourceLine: 1,
        RawLine: pasted,
        Name: `A\u0301ether 💫 ${'A'.repeat(90)}`,
        Reason: 'not found',
        Candidates: [{ Name: `💫${'B'.repeat(90)}`, Score: 1, LowConfidence: false }],
      }] }) },
    });
    const wrapper = mount(DeckImportPanel, { props: { sessionId: 'session-1' } });

    await wrapper.get('[data-testid="deck-text"]').setValue(pasted);
    await wrapper.get('form').trigger('submit');
    await flushPromises();

    expect(mutate().mock.calls[0][0].variables.input.text).toBe(pasted);
    const chip = wrapper.get('[data-testid="suggestion-chip"]');
    expect(chip.attributes('title')).toContain('💫');
    expect(chip.text()).not.toContain('\uFFFD');
  });

  it('discards an older response when two requests resolve out of order', async () => {
    const first = deferred<{ data: { previewDeck: ReturnType<typeof preview> } }>();
    const second = deferred<{ data: { previewDeck: ReturnType<typeof preview> } }>();
    mutate().mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
    const wrapper = mount(DeckImportPanel, { props: { sessionId: 'session-1', initialText: '1 Sol Ring' } });

    await wrapper.get('form').trigger('submit');
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    second.resolve({ data: { previewDeck: preview({ CardCount: 2 }) } });
    await flushPromises();
    first.resolve({ data: { previewDeck: preview({ CardCount: 99 }) } });
    await flushPromises();

    expect(wrapper.get('[data-testid="preview-totals"]').text()).toContain('2 of 2 cards ready');
    expect(wrapper.text()).not.toContain('99 of 99');
  });

  it('selects paste and URL inputs explicitly and always sends the session identifier', async () => {
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: preview() } })
      .mockResolvedValueOnce({ data: { previewDeck: preview({ SourceType: 'URL' }) } });
    const wrapper = mount(DeckImportPanel, { props: { sessionId: 'session-123' } });

    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sol Ring');
    await wrapper.get('form').trigger('submit');
    await flushPromises();
    expect(mutate().mock.calls[0][0].variables.input).toEqual({ text: '1 Sol Ring', sessionID: 'session-123' });

    await wrapper.get('[data-testid="source-url-tab"]').trigger('click');
    await wrapper.get('[data-testid="source-url"]').setValue('https://archidekt.com/decks/123');
    await wrapper.get('form').trigger('submit');
    await flushPromises();
    expect(mutate().mock.calls[1][0].variables.input).toEqual({
      sourceURL: 'https://archidekt.com/decks/123',
      sessionID: 'session-123',
    });
  });

  it('uses fixed-height wrapping input and capped overflow surfaces', () => {
    const wrapper = mount(DeckImportPanel, { props: { sessionId: 'session-1' } });
    expect(wrapper.get('[data-testid="deck-text"]').classes()).toContain('deck-textarea');
    expect(wrapper.get('[data-testid="deck-import-panel"]').classes()).toContain('deck-import-panel');
  });
});
