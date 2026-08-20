import { beforeEach, describe, expect, it } from 'vitest';

import {
  QUICK_START_DRAFT_KEY,
  clearDraft,
  readDraft,
  saveDraft,
  type QuickStartDraft,
} from '../src/services/quickStartDraft';

const draft: QuickStartDraft = {
  deckText: '1 A\u0301ether 💫 Adept',
  sourceURL: '',
  corrections: { 1: 'Æther Adept' },
  selectedCommanders: [{ ID: 'cmd-1', Name: 'Kraum, Ludevic’s Opus', Text: 'Partner' }],
  displayName: 'Zoë 💫',
};

beforeEach(() => {
  sessionStorage.clear();
});

describe('quickStartDraft', () => {
  it('round-trips every draft field with non-ASCII text unchanged', () => {
    saveDraft(draft);

    expect(readDraft()).toEqual(draft);
    expect(sessionStorage.getItem(QUICK_START_DRAFT_KEY)).toContain('Zoë 💫');
  });

  it.each(['{bad json', 'null', '[]', '{"deckText":3}', '{"deckText":"ok"}'])('discards malformed stored data: %s', (raw) => {
    sessionStorage.setItem(QUICK_START_DRAFT_KEY, raw);

    expect(readDraft()).toBeNull();
    expect(sessionStorage.getItem(QUICK_START_DRAFT_KEY)).toBeNull();
  });

  it('degrades to an absent draft when storage methods throw', () => {
    const original = globalThis.sessionStorage;
    Object.defineProperty(globalThis, 'sessionStorage', {
      configurable: true,
      value: {
        getItem() { throw new Error('storage denied'); },
        setItem() { throw new Error('storage denied'); },
        removeItem() { throw new Error('storage denied'); },
      },
    });

    expect(() => saveDraft(draft)).not.toThrow();
    expect(readDraft()).toBeNull();
    expect(() => clearDraft()).not.toThrow();

    Object.defineProperty(globalThis, 'sessionStorage', { configurable: true, value: original });
  });

  it('clears the draft explicitly', () => {
    saveDraft(draft);
    clearDraft();

    expect(readDraft()).toBeNull();
    expect(sessionStorage.getItem(QUICK_START_DRAFT_KEY)).toBeNull();
  });
});
