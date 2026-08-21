import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { flushPromises, mount } from '@vue/test-utils';

vi.mock('../src/services/apollo', () => ({
  apolloClient: { mutate: vi.fn(), query: vi.fn() },
}));

import { apolloClient } from '../src/services/apollo';
import DeckImportPanel from '../src/components/decks/DeckImportPanel.vue';
import {
  MAGIC_COMMANDER_DECK_CONTEXT,
  type DeckImportContext,
} from '../src/components/decks/deckImportContext';

type Mock = ReturnType<typeof vi.fn>;

function mutate(): Mock {
  return apolloClient.mutate as unknown as Mock;
}

function mountPanel(props: Partial<{
  sessionId: string;
  initialText: string;
  initialSourceURL: string;
  persistenceKey: string;
  context: DeckImportContext;
}> = {}) {
  return mount(DeckImportPanel, {
    props: {
      sessionId: 'session-1',
      context: MAGIC_COMMANDER_DECK_CONTEXT,
      ...props,
    },
  });
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
    SourceType: 'plain_text',
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
    const wrapper = mountPanel();

    expect(wrapper.text()).toContain('Paste your decklist');
    expect(wrapper.text()).toContain('Paste a deck export or plain-text list');
    expect(wrapper.text()).not.toContain('any format');
    expect(wrapper.get('[data-testid="deck-preview-submit"]').attributes('disabled')).toBeDefined();

    await (wrapper.vm as unknown as { submitPreview: () => Promise<void> }).submitPreview();
    expect(mutate()).not.toHaveBeenCalled();
    expect(wrapper.find('[data-testid="preview-totals"]').exists()).toBe(false);
  });

  it('visibly and accessibly identifies a typed, replaceable game and format context', () => {
    const futureContext = {
      gameLabel: 'Future card game',
      formatLabel: 'Future format',
    } satisfies DeckImportContext;
    const wrapper = mountPanel({ context: futureContext });
    const context = wrapper.get('[data-testid="deck-import-context"]');

    expect(context.findAll('dt').map((term) => term.text())).toEqual(['Game', 'Format']);
    expect(context.findAll('dd').map((value) => value.text())).toEqual([
      'Future card game',
      'Future format',
    ]);
    expect(wrapper.get('[data-testid="deck-import-panel"]').attributes('aria-label')).toBe(
      'Future card game Future format deck import',
    );
    expect(wrapper.text()).not.toContain('Magic: The Gathering');
    expect(wrapper.text()).not.toContain('Commander (EDH)');
  });

  it('marks the active source unmistakably and associates each tab with one exposed panel', async () => {
    const wrapper = mountPanel();
    const pasteTab = wrapper.get('[data-testid="paste-tab"]');
    const urlTab = wrapper.get('[data-testid="source-url-tab"]');
    const pastePanel = wrapper.get('[data-testid="paste-panel"]');
    const urlPanel = wrapper.get('[data-testid="source-url-panel"]');

    expect(wrapper.get('[role="tablist"]').attributes('aria-label')).toBe('Deck source');
    expect(pasteTab.attributes()).toMatchObject({
      role: 'tab',
      'aria-selected': 'true',
      'aria-controls': 'deck-source-paste-panel',
      tabindex: '0',
    });
    expect(pasteTab.classes()).toContain('is-selected');
    expect(pasteTab.get('.source-tab-marker').text()).toBe('✓');
    expect(urlTab.attributes()).toMatchObject({
      role: 'tab',
      'aria-selected': 'false',
      'aria-controls': 'deck-source-url-panel',
      tabindex: '-1',
    });
    expect(urlTab.classes()).not.toContain('is-selected');
    expect(pastePanel.attributes()).toMatchObject({
      role: 'tabpanel',
      'aria-labelledby': 'deck-source-paste-tab',
    });
    expect(pastePanel.attributes('hidden')).toBeUndefined();
    expect(urlPanel.attributes('hidden')).toBe('');

    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sol Ring');
    await urlTab.trigger('click');
    await wrapper.get('[data-testid="source-url"]').setValue('https://archidekt.com/decks/123');

    expect(pasteTab.attributes('aria-selected')).toBe('false');
    expect(pasteTab.attributes('tabindex')).toBe('-1');
    expect(pasteTab.classes()).not.toContain('is-selected');
    expect(pastePanel.attributes('hidden')).toBe('');
    expect(urlTab.attributes('aria-selected')).toBe('true');
    expect(urlTab.attributes('tabindex')).toBe('0');
    expect(urlTab.classes()).toContain('is-selected');
    expect(urlTab.get('.source-tab-marker').text()).toBe('✓');
    expect(urlPanel.attributes('hidden')).toBeUndefined();

    await pasteTab.trigger('click');
    expect((wrapper.get('[data-testid="deck-text"]').element as HTMLTextAreaElement).value).toBe('1 Sol Ring');
    await urlTab.trigger('click');
    expect((wrapper.get('[data-testid="source-url"]').element as HTMLInputElement).value).toBe(
      'https://archidekt.com/decks/123',
    );
  });

  it('switches and roves focus with tab keyboard controls', async () => {
    const host = document.createElement('div');
    document.body.append(host);
    const wrapper = mount(DeckImportPanel, {
      attachTo: host,
      props: {
        sessionId: 'session-1',
        context: MAGIC_COMMANDER_DECK_CONTEXT,
      },
    });
    const pasteTab = wrapper.get('[data-testid="paste-tab"]');
    const urlTab = wrapper.get('[data-testid="source-url-tab"]');

    (pasteTab.element as HTMLButtonElement).focus();
    await pasteTab.trigger('keydown', { key: 'ArrowRight' });
    await flushPromises();
    expect(urlTab.attributes('aria-selected')).toBe('true');
    expect(document.activeElement).toBe(urlTab.element);

    await urlTab.trigger('keydown', { key: 'Home' });
    await flushPromises();
    expect(pasteTab.attributes('aria-selected')).toBe('true');
    expect(document.activeElement).toBe(pasteTab.element);

    await pasteTab.trigger('keydown', { key: 'End' });
    await flushPromises();
    expect(urlTab.attributes('aria-selected')).toBe('true');
    expect(document.activeElement).toBe(urlTab.element);

    wrapper.unmount();
    host.remove();
  });

  it('describes the canonical paste grammar and only the registered URL provider', async () => {
    const wrapper = mountPanel();
    const pasteGuidance = wrapper.get('[data-testid="paste-format-guidance"]');
    const parserSource = readFileSync(resolve(process.cwd(), '../pkg/deckimport/scanner.go'), 'utf8');

    expect(parserSource).toContain('complete canonical decklist grammar: all six required syntaxes');
    for (const acceptedLine of ['1 Sol Ring', '1x Sol Ring', '1,Sol Ring', '1, Sol Ring']) {
      expect(parserSource).toContain(`\`${acceptedLine}\``);
      expect(pasteGuidance.text()).toContain(acceptedLine);
    }
    expect(pasteGuidance.text()).toContain('quoted CSV names');
    expect(pasteGuidance.text()).toContain('card name without a quantity');
    expect(wrapper.get('[data-testid="deck-text"]').attributes('aria-describedby')).toBe(
      'deck-text-format-guidance',
    );

    await wrapper.get('[data-testid="source-url-tab"]').trigger('click');
    const urlGuidance = wrapper.get('[data-testid="url-site-guidance"]');
    const providerSource = readFileSync(resolve(process.cwd(), '../server/deck_providers.go'), 'utf8');
    const registry = providerSource.match(
      /var deckProviderAdapters = map\[string\]deckProviderAdapter\{([\s\S]*?)\n\}/,
    )?.[1] ?? '';

    const registeredHosts = Array.from(registry.matchAll(/^[\t ]*(\w+Host):/gm), match => match[1]);
    expect(registeredHosts).toEqual(['archidektHost']);
    expect(providerSource).toContain('const archidektHost = "archidekt.com"');
    expect(urlGuidance.text()).toContain('Public Archidekt deck URLs are supported');
    expect(urlGuidance.text()).toContain('may be disabled by the server operator');
    expect(urlGuidance.text()).not.toContain('Moxfield');
    expect(wrapper.get('[data-testid="source-url"]').attributes('aria-describedby')).toBe(
      'deck-url-site-guidance',
    );
  });

  it('renders one panel-level loading state and disables submit while previewDeck is pending', async () => {
    const pending = deferred<{ data: { previewDeck: ReturnType<typeof preview> } }>();
    mutate().mockReturnValueOnce(pending.promise);
    const wrapper = mountPanel();

    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sol Ring');
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');

    expect(wrapper.text()).toContain('Checking your deck');
    expect(wrapper.find('[data-testid="preview-spinner"]').exists()).toBe(true);
    expect(wrapper.get('[data-testid="deck-preview-submit"]').attributes('disabled')).toBeDefined();

    pending.resolve({ data: { previewDeck: preview() } });
    await flushPromises();
    expect(wrapper.find('[data-testid="preview-spinner"]').exists()).toBe(false);
  });

  it('acknowledges a ready paste import with useful, accessible success context', async () => {
    mutate().mockResolvedValueOnce({ data: { previewDeck: preview({ CardCount: 100 }) } });
    const wrapper = mountPanel({ initialText: '1 Sol Ring' });

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();

    const acknowledgement = wrapper.get('[data-testid="paste-import-success"]');
    expect(acknowledgement.attributes()).toMatchObject({
      role: 'status',
      'aria-live': 'polite',
      'aria-atomic': 'true',
    });
    expect(acknowledgement.text()).toContain('Deck import successful');
    expect(acknowledgement.text()).toContain('100 cards are ready');
    expect(acknowledgement.text()).toContain('Continue to choose your commander');
    expect(acknowledgement.get('.paste-import-success-icon').text()).toBe('✓');
    expect(acknowledgement.get('.paste-import-success-icon').attributes('aria-hidden')).toBe('true');
    expect(wrapper.get('[data-testid="preview-totals"]').attributes('aria-live')).toBe('off');
  });

  it.each([
    ['server says the deck cannot continue', preview({ CanContinue: false })],
    ['blocking errors remain', preview({ BlockingErrors: ['blocked'], CanContinue: true })],
  ])('does not acknowledge success when %s', async (_case, result) => {
    mutate().mockResolvedValueOnce({ data: { previewDeck: result } });
    const wrapper = mountPanel({ initialText: '1 Sol Ring' });

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();

    expect(wrapper.find('[data-testid="paste-import-success"]').exists()).toBe(false);
  });

  it('does not acknowledge a URL-tab import even when the provider response is actionable', async () => {
    mutate().mockResolvedValueOnce({
      data: { previewDeck: preview({ SourceType: 'archidekt' }) },
    });
    const wrapper = mountPanel({ initialSourceURL: 'https://archidekt.com/decks/123' });

    expect(wrapper.get('[data-testid="source-url-tab"]').attributes('aria-selected')).toBe('true');
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();

    expect(wrapper.find('[data-testid="paste-import-success"]').exists()).toBe(false);
  });

  it('never acknowledges loading or transport failure states', async () => {
    const pending = deferred<{ data: { previewDeck: ReturnType<typeof preview> } }>();
    mutate().mockReturnValueOnce(pending.promise).mockRejectedValueOnce(
      apolloError('preview_error', 'raw transport failure'),
    );
    const wrapper = mountPanel({ initialText: '1 Sol Ring' });

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    expect(wrapper.find('[data-testid="paste-import-success"]').exists()).toBe(false);
    pending.resolve({ data: { previewDeck: preview() } });
    await flushPromises();
    expect(wrapper.find('[data-testid="paste-import-success"]').exists()).toBe(true);

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    expect(wrapper.find('[data-testid="paste-import-success"]').exists()).toBe(false);
    expect(wrapper.find('[data-testid="activation-error"]').exists()).toBe(true);
  });

  it('acknowledges the transition after every unresolved card receives an explicit correction', async () => {
    const unresolved = [
      {
        SourceLine: 1,
        RawLine: '1 Sl Ring',
        Name: 'Sl Ring',
        Reason: 'not found',
        Candidates: [{ Name: 'Sol Ring', Score: 0.98, LowConfidence: false }],
      },
      {
        SourceLine: 2,
        RawLine: '1 Arcane Sign',
        Name: 'Arcane Sign',
        Reason: 'not found',
        Candidates: [{ Name: 'Arcane Signet', Score: 0.96, LowConfidence: false }],
      },
    ];
    mutate().mockResolvedValueOnce({
      data: { previewDeck: preview({ CardCount: 100, Unresolved: unresolved }) },
    });
    const wrapper = mountPanel({ initialText: '1 Sl Ring\n1 Arcane Sign' });

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    expect(wrapper.find('[data-testid="paste-import-success"]').exists()).toBe(false);
    expect(wrapper.get('[data-testid="preview-totals"]').attributes('aria-live')).toBe('polite');
    expect(wrapper.get('[data-testid="preview-totals"]').text()).toContain('2 cards need a quick look');

    const suggestions = wrapper.findAll('[data-testid="suggestion-chip"]');
    await suggestions[0].trigger('click');
    expect(wrapper.find('[data-testid="paste-import-success"]').exists()).toBe(false);
    expect(wrapper.get('[data-testid="preview-totals"]').text()).toContain('1 card needs a quick look');

    await suggestions[1].trigger('click');
    await flushPromises();
    const acknowledgement = wrapper.get('[data-testid="paste-import-success"]');
    const acknowledgementElement = acknowledgement.element;
    expect(wrapper.get('[data-testid="preview-totals"]').text()).toContain('100 of 100 cards ready');
    expect(acknowledgement.text()).toContain('2 corrections applied');

    await suggestions[1].trigger('click');
    await flushPromises();
    expect(wrapper.get('[data-testid="paste-import-success"]').element).toBe(acknowledgementElement);
  });

  it('resets during a subsequent import and acknowledges the newly ready result once', async () => {
    const next = deferred<{ data: { previewDeck: ReturnType<typeof preview> } }>();
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: preview({ CardCount: 1 }) } })
      .mockReturnValueOnce(next.promise);
    const wrapper = mountPanel({ initialText: '1 Sol Ring' });

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    const firstElement = wrapper.get('[data-testid="paste-import-success"]').element;

    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sol Ring\n1 Arcane Signet');
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    expect(wrapper.find('[data-testid="paste-import-success"]').exists()).toBe(false);

    next.resolve({ data: { previewDeck: preview({ CardCount: 2 }) } });
    await flushPromises();
    const second = wrapper.get('[data-testid="paste-import-success"]');
    expect(second.element).not.toBe(firstElement);
    expect(second.text()).toContain('2 cards are ready');
    expect(wrapper.findAll('[data-testid="paste-import-success"]')).toHaveLength(1);
  });

  it.each([
    [0, 'No warnings'],
    [1, '1 warning'],
    [3, '3 warnings'],
  ])('pluralizes %i warnings in a resolved totals card', async (count, copy) => {
    mutate().mockResolvedValueOnce({ data: { previewDeck: preview({ Warnings: Array(count).fill('warning') }) } });
    const wrapper = mountPanel({ initialText: '1 Sol Ring' });

    expect(wrapper.find('[data-testid="preview-totals"]').exists()).toBe(false);
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
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
    const wrapper = mountPanel({ initialText: '1 Sol Ring' });

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
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
    const wrapper = mountPanel({ initialText: '1 Typo Card' });

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    expect(wrapper.find('[data-testid="unresolved-list"]').exists()).toBe(false);

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    expect(wrapper.text()).toContain('1 card needs a quick look');
    expect(wrapper.findAll('[data-testid="suggestion-chip"]')).toHaveLength(3);
    expect(wrapper.find('[data-testid="manual-card-name"]').exists()).toBe(true);
    expect(wrapper.find('[data-testid="remove-unresolved-line"]').exists()).toBe(true);

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
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
    const wrapper = mountPanel({ initialText: '1 Sl Ring' });

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    expect(wrapper.emitted('change')?.at(-1)?.[0]).toMatchObject({ corrections: {} });

    await wrapper.get('[data-testid="suggestion-chip"]').trigger('click');
    expect(wrapper.emitted('change')?.at(-1)?.[0]).toMatchObject({ corrections: { 1: 'Sol Ring' } });

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
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
    const wrapper = mountPanel({ initialText: '1 Sol Ring' });

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
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
    const wrapper = mountPanel();

    await wrapper.get('[data-testid="deck-text"]').setValue(pasted);
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
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
    const wrapper = mountPanel({ initialText: '1 Sol Ring' });

    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    // A disabled button prevents a second user click, while a programmatic
    // submit (or two submits dispatched before the DOM reflects loading)
    // still exercises the component's stale-response safety boundary.
    const secondSubmit = (wrapper.vm as unknown as { submitPreview: () => Promise<void> }).submitPreview();
    second.resolve({ data: { previewDeck: preview({ CardCount: 2 }) } });
    await flushPromises();
    first.resolve({ data: { previewDeck: preview({ CardCount: 99 }) } });
    await secondSubmit;
    await flushPromises();

    expect(wrapper.get('[data-testid="preview-totals"]').text()).toContain('2 of 2 cards ready');
    expect(wrapper.text()).not.toContain('99 of 99');
  });

  it('selects paste and URL inputs explicitly and always sends the session identifier', async () => {
    mutate()
      .mockResolvedValueOnce({ data: { previewDeck: preview() } })
      .mockResolvedValueOnce({ data: { previewDeck: preview({ SourceType: 'URL' }) } });
    const wrapper = mountPanel({ sessionId: 'session-123' });

    await wrapper.get('[data-testid="deck-text"]').setValue('1 Sol Ring');
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    expect(mutate().mock.calls[0][0].variables.input).toEqual({ text: '1 Sol Ring', sessionID: 'session-123' });

    await wrapper.get('[data-testid="source-url-tab"]').trigger('click');
    await wrapper.get('[data-testid="source-url"]').setValue('https://archidekt.com/decks/123');
    await wrapper.get('[data-testid="deck-preview-submit"]').trigger('click');
    await flushPromises();
    expect(mutate().mock.calls[1][0].variables.input).toEqual({
      sourceURL: 'https://archidekt.com/decks/123',
      sessionID: 'session-123',
    });
  });

  it('uses fixed-height wrapping input and capped overflow surfaces', () => {
    const wrapper = mountPanel();
    const panelSource = readFileSync(resolve(process.cwd(), 'src/components/decks/DeckImportPanel.vue'), 'utf8');
    const quickStartSource = readFileSync(resolve(process.cwd(), 'src/views/QuickStartView.vue'), 'utf8');

    expect(wrapper.get('[data-testid="deck-text"]').classes()).toContain('deck-textarea');
    expect(wrapper.get('[data-testid="deck-import-panel"]').classes()).toContain('deck-import-panel');
    expect(panelSource).toContain('grid-template-columns: minmax(0, 1fr)');
    expect(panelSource).toContain('grid-template-columns: repeat(2, minmax(0, 1fr))');
    expect(panelSource).toContain('.source-panel[hidden]');
    expect(panelSource).toContain('@keyframes paste-import-success-arrive');
    expect(panelSource).toContain('@media (prefers-reduced-motion: reduce)');
    expect(panelSource).toMatch(/prefers-reduced-motion: reduce[\s\S]*\.paste-import-success[\s\S]*animation: none/);
    expect(panelSource).toContain('@media (min-width: bp.$breakpoint-tablet)');
    expect(panelSource).toContain('grid-template-columns: minmax(0, 3fr) minmax(0, 2fr)');
    expect(quickStartSource).toContain("'quick-start-deck-import--wide': stage === 'import'");
    expect(quickStartSource).toContain('grid-column: 1 / -1');
  });
});
