import { describe, it, expect } from 'vitest';
import {
  hasPartnerAbility,
  isValidPartnerPair,
  partnerWithTargetName,
  partnerConstraintMessage,
} from '../src/services/commanderPartner';

describe('commanderPartner helpers', () => {
  it('detects Partner keyword', () => {
    expect(hasPartnerAbility({ ID: '1', Name: 'A', Text: 'Partner' })).toBe(true);
    expect(hasPartnerAbility({ ID: '1', Name: 'A', Text: 'partner with B (When...)' })).toBe(true);
    expect(hasPartnerAbility({ ID: '1', Name: 'A', Text: 'No relevant keyword' })).toBe(false);
  });

  it('parses Partner with target name', () => {
    expect(partnerWithTargetName({ ID: '1', Name: 'A', Text: 'Partner with Toothy, Imaginary Friend (When...)' })).toBe(
      'Toothy, Imaginary Friend',
    );
    expect(partnerWithTargetName({ ID: '1', Name: 'A', Text: 'Partner' })).toBe(null);
  });

  it('validates Partner pairs', () => {
    const a = { ID: 'a', Name: 'A', Text: 'Partner' };
    const b = { ID: 'b', Name: 'B', Text: 'Partner' };
    expect(isValidPartnerPair(a, b)).toBe(true);

    const notPartner = { ID: 'c', Name: 'C', Text: 'Legendary Creature' };
    expect(isValidPartnerPair(a, notPartner)).toBe(false);
  });

  it('validates Partner with constraints', () => {
    const p1 = { ID: 'a', Name: 'Pir, Imaginative Rascal', Text: 'Partner with Toothy, Imaginary Friend (When...)' };
    const toothy = { ID: 'b', Name: 'Toothy, Imaginary Friend', Text: 'Partner with Pir, Imaginative Rascal (When...)' };
    const wrong = { ID: 'c', Name: 'B', Text: 'Partner' };

    expect(isValidPartnerPair(p1, toothy)).toBe(true);
    expect(isValidPartnerPair(p1, wrong)).toBe(false);
    expect(partnerConstraintMessage(p1)).toMatch('Second commander must be');
  });
});
