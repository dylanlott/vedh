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
 * Fire-and-forget product-event emission (D-19). Sends exactly one mutation
 * per call — never batched, never awaited by the caller, and returns
 * undefined synchronously. A rejected mutation is swallowed: logged at debug
 * level under the [productEvents] prefix, never surfaced as an unhandled
 * rejection or at console.error level. Never reads the auth storage key and
 * never places an authorization value into metadata; the server attaches
 * the authenticated user identifier from the request context.
 */
export function track(name: string, metadata: Record<string, string | number> = {}): void {
  const metadataList: InputProductEventMeta[] = Object.entries(metadata).map(([key, value]) => ({
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
