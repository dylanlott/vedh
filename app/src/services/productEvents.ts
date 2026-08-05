// Browser-side product-event emission. Plain module, named exports only, no
// Pinia, no default export — mirrors app/src/services/commanderPartner.ts's
// shape.
//
// The per-event attribution table below (added in a later task) is a
// convenience, not a control: the server-side per-event metadata key
// allowlist in pkg/telemetry.Vocabulary (see
// docs/analytics/product-event-vocabulary.md) is the enforcement point,
// because the client is not a trust boundary. A mismatch here causes the
// server to drop the event and increment vedh_product_events_dropped_total,
// which is visible and safe, rather than to store something it should not.

import { apolloClient } from './apollo';
import { TRACK_PRODUCT_EVENT_MUTATION } from '../graphql/mutations';
import type { InputProductEventMeta } from '../types/generated';

const STORAGE_KEY = 'edhgo/session-id';

/** Byte (not character) length limit for any single attribution value, matching the server's MaxMetadataValueBytes. */
const MAX_ATTRIBUTION_VALUE_BYTES = 128;

/**
 * The campaign-attribution keys read out of the page query string. This is a
 * strict subset of the four attribution keys
 * docs/analytics/product-event-vocabulary.md records for the events that
 * accept attribution (those four are these three plus `referrer_host`,
 * which is derived from the referrer below, never from the query string).
 * Frozen so this allowlist cannot be mutated at runtime.
 */
const QUERY_ATTRIBUTION_KEYS = Object.freeze(['utm_source', 'utm_medium', 'utm_campaign'] as const);

/** The metadata key the referrer's host is captured under. */
const REFERRER_ATTRIBUTION_KEY = 'referrer_host';

/**
 * Per-event decision table mirroring docs/analytics/product-event-vocabulary.md:
 * of the 15 events in the closed vocabulary, only `landing_primary_cta` and
 * `invite_viewed` allowlist the four attribution keys server-side. This
 * table is a convenience, not a control — the server-side per-event
 * allowlist in pkg/telemetry.Vocabulary is what actually enforces it. A
 * mismatch here causes the server to drop the event and increment its
 * bounded vedh_product_events_dropped_total{reason} counter, which is
 * visible and safe, rather than to store something it should not.
 */
const EVENTS_ACCEPTING_ATTRIBUTION: ReadonlySet<string> = new Set(['landing_primary_cta', 'invite_viewed']);

/**
 * Attribution captured once per page view and reused for every subsequent
 * track() call, rather than re-read from the URL each time.
 */
let cachedAttribution: Record<string, string> | null = null;

/**
 * Defensive storage accessor mirroring app/src/services/apollo.ts's
 * getRawAuth: try the direct global first, then the window-scoped global,
 * and return null rather than propagating. Private-browsing mode and a
 * disabled-storage configuration must degrade, never throw.
 */
function getStorage(): Storage | null {
  try {
    if (typeof localStorage !== 'undefined' && localStorage) return localStorage;
  } catch {
    // ignore — fall through to the window-scoped attempt
  }
  try {
    if (typeof window !== 'undefined' && window.localStorage) return window.localStorage;
  } catch {
    // ignore — no storage available in this environment
  }
  return null;
}

/**
 * Generates a fresh session identifier. Tries crypto.randomUUID() first
 * (secure contexts only — a LAN-IP dev server over plain HTTP, a normal way
 * to test on a tablet, leaves this undefined), then falls back to
 * crypto.getRandomValues() hex-encoded into a 32-character string. The final
 * branch is a timestamp-and-pseudorandom composition that exists only so
 * this service degrades rather than throwing when neither cryptographic
 * source is present; it is never relied upon for unpredictability.
 */
function newSessionID(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID();
  }
  if (typeof crypto !== 'undefined' && typeof crypto.getRandomValues === 'function') {
    const bytes = new Uint8Array(16);
    crypto.getRandomValues(bytes);
    return Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('');
  }
  return `${Date.now().toString(16)}${Math.random().toString(16).slice(2)}`;
}

/**
 * Returns the persisted per-browser session identifier (D-18), generating
 * and persisting one on first use. Stable across tabs, reloads, and
 * restarts. Never throws: any storage failure degrades to a freshly
 * generated, unpersisted identifier.
 */
export function getSessionID(): string {
  try {
    const storage = getStorage();
    const existing = storage?.getItem(STORAGE_KEY);
    if (existing) return existing;
    const fresh = newSessionID();
    storage?.setItem(STORAGE_KEY, fresh);
    return fresh;
  } catch {
    return newSessionID();
  }
}

/**
 * Guarded accessor for the page location, in the same defensive shape as
 * getStorage(): the `window`/`location` globals are also absent in some
 * environments (e.g. non-DOM test harnesses), so this never throws.
 */
function getLocationSearch(): string {
  try {
    if (typeof window !== 'undefined' && window.location) return window.location.search ?? '';
  } catch {
    // ignore — no location available in this environment
  }
  return '';
}

/** Guarded accessor for document.referrer, never throwing. */
function getReferrer(): string {
  try {
    if (typeof document !== 'undefined' && document.referrer) return document.referrer;
  } catch {
    // ignore — no document available in this environment
  }
  return '';
}

/**
 * Truncates `value` to at most `maxBytes` bytes, measured as raw UTF-8
 * bytes via TextEncoder — never as character count — matching the server's
 * limit. Truncation always lands on a codepoint boundary (Array.from splits
 * a string into whole codepoints, respecting surrogate pairs), so a
 * multi-byte sequence is never cut in half.
 */
function truncateToByteLimit(value: string, maxBytes: number): string {
  const encoder = new TextEncoder();
  if (encoder.encode(value).length <= maxBytes) return value;

  let codepoints = Array.from(value);
  while (codepoints.length > 0) {
    const candidate = codepoints.join('');
    if (encoder.encode(candidate).length <= maxBytes) return candidate;
    codepoints = codepoints.slice(0, -1);
  }
  return '';
}

/**
 * Captures allowlisted campaign attribution from the page URL and referrer
 * (T-01-17). Reads the query string once per module lifetime and memoises
 * the result so repeated track() calls in one page view share the same
 * values rather than re-parsing the URL. Every value is trimmed and
 * truncated to at most MAX_ATTRIBUTION_VALUE_BYTES bytes on a codepoint
 * boundary; a key whose value is empty after trimming is dropped rather
 * than emitted as an empty string. The referrer is reduced to its
 * lower-cased host only — never its path, query string, or fragment, which
 * could carry a search term or session token.
 */
export function captureAttribution(): Record<string, string> {
  if (cachedAttribution) return cachedAttribution;

  const result: Record<string, string> = {};

  try {
    const search = getLocationSearch();
    if (search) {
      const params = new URLSearchParams(search);
      for (const key of QUERY_ATTRIBUTION_KEYS) {
        const raw = params.get(key);
        if (raw === null) continue;
        const trimmed = raw.trim();
        if (!trimmed) continue;
        result[key] = truncateToByteLimit(trimmed, MAX_ATTRIBUTION_VALUE_BYTES);
      }
    }
  } catch {
    // A malformed query string degrades to no attribution, never a throw.
  }

  try {
    const referrer = getReferrer();
    if (referrer) {
      const host = new URL(referrer).host.trim().toLowerCase();
      if (host) result[REFERRER_ATTRIBUTION_KEY] = truncateToByteLimit(host, MAX_ATTRIBUTION_VALUE_BYTES);
    }
  } catch {
    // An unparsable referrer degrades to no referrer_host, never a throw.
  }

  cachedAttribution = result;
  return result;
}

/**
 * Fire-and-forget product-event emission (D-19). Sends exactly one mutation
 * per call — never batched, never awaited by the caller, and returns
 * undefined synchronously. A rejected mutation is swallowed: logged at debug
 * level under the [productEvents] prefix, never surfaced as an unhandled
 * rejection or at console.error level. Never reads the auth storage key and
 * never places an authorization value into metadata; the server attaches
 * the authenticated user identifier from the request context.
 */
export function track(name: string, metadata: Record<string, string | number> = {}): void {
  const attribution = EVENTS_ACCEPTING_ATTRIBUTION.has(name) ? captureAttribution() : {};
  const merged: Record<string, string | number> = { ...attribution, ...metadata };
  const metadataList: InputProductEventMeta[] = Object.entries(merged).map(([key, value]) => ({
    key,
    value: String(value),
  }));

  void apolloClient
    .mutate({
      mutation: TRACK_PRODUCT_EVENT_MUTATION,
      variables: {
        input: {
          name,
          sessionID: getSessionID(),
          metadata: metadataList,
        },
      },
    })
    .catch((err) => {
      // eslint-disable-next-line no-console
      console.debug('[productEvents] drop', name, err);
    });
}
