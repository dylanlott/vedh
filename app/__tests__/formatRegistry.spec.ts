import { describe, expect, it } from 'vitest';
import { formats, lookupFormat } from '../src/formats/registry';

describe('formatRegistry', () => {
  it('exposes commander and generic duel presets', () => {
    expect(formats.map(format => format.ID)).toEqual(['EDH', 'GENERIC_DUEL']);
    expect(lookupFormat('GENERIC_DUEL').DefaultDeckSize).toBe(60);
    expect(lookupFormat('EDH').CommanderEnabled).toBe(true);
  });
});
