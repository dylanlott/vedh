import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// The Apollo client is mocked at the module boundary so this suite never
// makes a real network call. Each test that needs a fresh module instance
// (to prove storage-backed stability across a "reload", or attribution
// caching within one "page view") calls loadService(), which resets the
// module registry and re-imports both the mock and the service under test.
vi.mock('../src/services/apollo', () => ({
  apolloClient: { mutate: vi.fn() },
}));

type MutateMock = ReturnType<typeof vi.fn>;

async function loadService() {
  vi.resetModules();
  const apolloModule = await import('../src/services/apollo');
  const service = await import('../src/services/productEvents');
  const mutate = apolloModule.apolloClient.mutate as unknown as MutateMock;
  mutate.mockReset();
  mutate.mockResolvedValue(true);
  return { service, mutate };
}

const STORAGE_KEY = 'edhgo/session-id';
const AUTH_STORAGE_KEY = 'edhgo/auth';

function setLocationSearch(search: string) {
  const url = `http://localhost/${search ? `?${search}` : ''}`;
  window.history.pushState({}, '', url);
}

function setReferrer(value: string) {
  Object.defineProperty(document, 'referrer', { value, configurable: true });
}

beforeEach(() => {
  localStorage.clear();
  setLocationSearch('');
  setReferrer('');
});

afterEach(() => {
  vi.restoreAllMocks();
  localStorage.clear();
});

describe('getSessionID', () => {
  it('returns the identical string across two successive calls', async () => {
    const { service } = await loadService();
    const first = service.getSessionID();
    const second = service.getSessionID();
    expect(first).toBe(second);
    expect(first.length).toBeGreaterThan(0);
  });

  it('returns the same string across a simulated reload against the same storage', async () => {
    const { service: before } = await loadService();
    const original = before.getSessionID();

    const { service: after } = await loadService();
    const reloaded = after.getSessionID();

    expect(reloaded).toBe(original);
  });

  it('creates exactly one storage entry under edhgo/session-id', async () => {
    const setItemSpy = vi.spyOn(Storage.prototype, 'setItem');
    const { service } = await loadService();

    service.getSessionID();
    service.getSessionID();

    const sessionIdWrites = setItemSpy.mock.calls.filter(([key]) => key === STORAGE_KEY);
    expect(sessionIdWrites).toHaveLength(1);
  });

  it('returns a non-empty string and throws nothing when the storage getter throws', async () => {
    const { service } = await loadService();
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('storage disabled');
    });

    let result = '';
    expect(() => {
      result = service.getSessionID();
    }).not.toThrow();
    expect(result.length).toBeGreaterThan(0);
  });

  it('generates and persists a fresh identifier when the stored value is an empty string', async () => {
    localStorage.setItem(STORAGE_KEY, '');
    const { service } = await loadService();

    const id = service.getSessionID();

    expect(id.length).toBeGreaterThan(0);
    expect(localStorage.getItem(STORAGE_KEY)).toBe(id);
  });

  it('falls back to the random-bytes API when randomUUID is absent from the crypto global', async () => {
    const originalCrypto = globalThis.crypto;
    Object.defineProperty(globalThis, 'crypto', {
      value: { getRandomValues: originalCrypto.getRandomValues.bind(originalCrypto) },
      configurable: true,
    });

    try {
      const { service } = await loadService();
      const id = service.getSessionID();
      expect(id).toMatch(/^[0-9a-f]{32}$/);
    } finally {
      Object.defineProperty(globalThis, 'crypto', { value: originalCrypto, configurable: true });
    }
  });
});

describe('track', () => {
  it('returns undefined synchronously, not a thenable', async () => {
    const { service } = await loadService();
    const result = service.track('quick_start_viewed');
    expect(result).toBeUndefined();
  });

  it('calls mutate exactly once per invocation, with distinct variables and no batching', async () => {
    const { service, mutate } = await loadService();

    service.track('quick_start_viewed', { a: 1 });
    service.track('board_ready', { b: 2 });

    expect(mutate).toHaveBeenCalledTimes(2);
    expect(mutate.mock.calls[0][0]).not.toEqual(mutate.mock.calls[1][0]);
  });

  it('carries the session identifier from getSessionID() on every mutate call', async () => {
    const { service, mutate } = await loadService();
    const sessionID = service.getSessionID();

    service.track('quick_start_viewed');

    expect(mutate.mock.calls[0][0].variables.input.sessionID).toBe(sessionID);
  });

  it('swallows a rejected mutation without throwing, without an unhandled rejection, and without console.error', async () => {
    const { service, mutate } = await loadService();
    mutate.mockRejectedValue(new Error('network down'));
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    expect(() => service.track('quick_start_viewed')).not.toThrow();
    // Flush the microtask queue so the attached .catch handler runs before
    // we assert nothing escalated to console.error.
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(errorSpy).not.toHaveBeenCalled();
  });

  it('never reads the edhgo/auth storage key', async () => {
    const getItemSpy = vi.spyOn(Storage.prototype, 'getItem');
    const { service } = await loadService();

    service.track('quick_start_viewed');

    const keysRead = new Set(getItemSpy.mock.calls.map(([key]) => key));
    expect(keysRead.has(AUTH_STORAGE_KEY)).toBe(false);
  });
});

describe('captureAttribution', () => {
  it('returns all four allowlisted entries when the query string carries the three UTM keys and a referrer is present', async () => {
    const params = new URLSearchParams({
      utm_source: 'newsletter',
      utm_medium: 'email',
      utm_campaign: 'spring-launch',
    });
    setLocationSearch(params.toString());
    setReferrer('https://example.com/some/path?x=1');

    const { service } = await loadService();
    const attribution = service.captureAttribution();

    expect(attribution).toEqual({
      utm_source: 'newsletter',
      utm_medium: 'email',
      utm_campaign: 'spring-launch',
      referrer_host: 'example.com',
    });
  });

  it('omits a query-string key outside the allowlist', async () => {
    const params = new URLSearchParams({ utm_source: 'newsletter', foo: 'bar' });
    setLocationSearch(params.toString());

    const { service } = await loadService();
    const attribution = service.captureAttribution();

    expect(attribution).not.toHaveProperty('foo');
    expect(attribution.utm_source).toBe('newsletter');
  });

  it('truncates a value longer than 128 bytes on a character boundary', async () => {
    // Each '€' is a 3-byte UTF-8 sequence; 50 of them is 150 bytes.
    const longValue = '€'.repeat(50);
    const params = new URLSearchParams({ utm_campaign: longValue });
    setLocationSearch(params.toString());

    const { service } = await loadService();
    const attribution = service.captureAttribution();

    const byteLength = new TextEncoder().encode(attribution.utm_campaign).length;
    expect(byteLength).toBeLessThanOrEqual(128);
    // A valid truncation never produces the U+FFFD replacement character,
    // which is what a split multi-byte sequence would decode to.
    expect(attribution.utm_campaign).not.toContain('�');
  });

  it('trims surrounding whitespace from a value', async () => {
    const params = new URLSearchParams({ utm_source: '  padded  ' });
    setLocationSearch(params.toString());

    const { service } = await loadService();
    const attribution = service.captureAttribution();

    expect(attribution.utm_source).toBe('padded');
  });

  it('reduces a full referrer URL with path, query string, and fragment to its lower-cased host only', async () => {
    setReferrer('https://Example.COM/some/path?secret=token#frag');

    const { service } = await loadService();
    const attribution = service.captureAttribution();

    expect(attribution.referrer_host).toBe('example.com');
    expect(Object.values(attribution).join(' ')).not.toContain('secret');
    expect(Object.values(attribution).join(' ')).not.toContain('frag');
  });

  it('emits no referrer entry at all when the referrer is empty', async () => {
    setReferrer('');

    const { service } = await loadService();
    const attribution = service.captureAttribution();

    expect(attribution).not.toHaveProperty('referrer_host');
  });

  it('returns an empty object with no query string and no referrer, and track() still emits the event', async () => {
    const { service, mutate } = await loadService();
    const attribution = service.captureAttribution();

    expect(attribution).toEqual({});

    service.track('landing_primary_cta');
    expect(mutate).toHaveBeenCalledTimes(1);
  });

  it('reuses attribution captured once for subsequent track() calls rather than re-reading the URL', async () => {
    const params = new URLSearchParams({ utm_source: 'first' });
    setLocationSearch(params.toString());
    const { service, mutate } = await loadService();

    service.track('landing_primary_cta');
    setLocationSearch(new URLSearchParams({ utm_source: 'second' }).toString());
    service.track('landing_primary_cta');

    const firstMetadata = mutate.mock.calls[0][0].variables.input.metadata;
    const secondMetadata = mutate.mock.calls[1][0].variables.input.metadata;
    expect(firstMetadata).toEqual(secondMetadata);
    expect(firstMetadata.find((m: { key: string }) => m.key === 'utm_source').value).toBe('first');
  });

  it('attaches attribution only to events the vocabulary marks as accepting it', async () => {
    const params = new URLSearchParams({ utm_source: 'newsletter' });
    setLocationSearch(params.toString());
    const { service, mutate } = await loadService();

    service.track('quick_start_viewed');
    service.track('landing_primary_cta');

    const quickStartMetadata = mutate.mock.calls[0][0].variables.input.metadata;
    const landingMetadata = mutate.mock.calls[1][0].variables.input.metadata;

    expect(quickStartMetadata.some((m: { key: string }) => m.key === 'utm_source')).toBe(false);
    expect(landingMetadata.some((m: { key: string }) => m.key === 'utm_source')).toBe(true);
  });
});
