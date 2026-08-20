import { describe, expect, it } from 'vitest';

import {
  ACTIVATION_ERRORS,
  GENERIC_ACTIVATION_ERROR,
  resolveActivationError,
} from '../src/services/activationErrors';

describe('activation error copy contract', () => {
  it.each([
    [
      'provider_unavailable',
      "We can't reach that deck link right now.",
      'Paste your decklist instead — it takes about 10 seconds.',
      'Paste instead',
    ],
    [
      'preview_error',
      "We couldn't fully read a few cards.",
      'Fix the highlighted entries below, then continue — your original paste is untouched.',
      'Fix these cards',
    ],
    [
      'guest_session_error',
      "We couldn't set up your seat at the table.",
      'Nothing you entered was lost. Try again in a moment.',
      'Try again',
    ],
    [
      'create_error',
      "We couldn't create your table.",
      'Your deck and commander picks are still here — try again.',
      'Try again',
    ],
  ] as const)('maps %s to one fixed copy block', (code, title, body, affordance) => {
    const rawMessage = `raw-${code}-SQL-provider-secret`;
    const result = resolveActivationError({
      message: rawMessage,
      graphQLErrors: [{ message: rawMessage, extensions: { code } }],
    });

    expect(result).toBe(ACTIVATION_ERRORS[code]);
    expect(result).toEqual({ title, body, affordance });
    expect(JSON.stringify(result)).not.toContain(rawMessage);
  });

  it.each([
    new Error('raw transport details'),
    { graphQLErrors: [] },
    { graphQLErrors: [{ extensions: {} }] },
    { graphQLErrors: [{ extensions: { code: 'unknown_code' } }] },
    null,
  ])('uses one generic entry for unknown or malformed errors', (error) => {
    expect(resolveActivationError(error)).toBe(GENERIC_ACTIVATION_ERROR);
    expect(JSON.stringify(resolveActivationError(error))).not.toContain('raw transport details');
  });

  it('exports immutable code and fallback records', () => {
    expect(Object.isFrozen(ACTIVATION_ERRORS)).toBe(true);
    expect(Object.isFrozen(GENERIC_ACTIVATION_ERROR)).toBe(true);
    expect(Object.values(ACTIVATION_ERRORS).every(Object.isFrozen)).toBe(true);
  });
});
