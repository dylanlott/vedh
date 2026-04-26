import { describe, it, expect } from 'vitest';
import { isLandCard, moveHandCardToStackState, resolveStackCardToGraveyardState } from '../src/utils/stack';

describe('stack utils', () => {
  it('detects land cards by type', () => {
    expect(isLandCard({ Name: 'Forest', Types: 'Basic Land — Forest' })).toBe(true);
    expect(isLandCard({ Name: 'Shock', Types: 'Instant' })).toBe(false);
    expect(isLandCard(undefined)).toBe(false);
  });

  it('moves a hand card to stack with owner and prevents duplicates', () => {
    const hand = [
      { ID: 'a', Name: 'Shock', Types: 'Instant' },
      { ID: 'b', Name: 'Island', Types: 'Basic Land — Island' },
    ];
    const stack = [{ ID: 'x', Name: 'Opt' }];
    const moved = moveHandCardToStackState(hand, stack, hand[0], 'alice', 0);
    expect(moved.hand).toHaveLength(1);
    expect(moved.stack).toHaveLength(2);
    expect(moved.movedCard?.CurrentZone).toBe('alice');
    expect(moved.movedCard?.Tapped).toBe(false);

    const dup = moveHandCardToStackState(moved.hand, moved.stack, hand[0], 'alice', 0);
    expect(dup.skippedReason).toBe('duplicate');
  });

  it('resolves stack card into graveyard by index', () => {
    const stack = [{ ID: 'a', Name: 'Shock' }, { ID: 'b', Name: 'Opt' }];
    const graveyard = [{ ID: 'c', Name: 'Bolt' }];
    const resolved = resolveStackCardToGraveyardState(stack, graveyard, 0);
    expect(resolved.stack).toHaveLength(1);
    expect(resolved.graveyard).toHaveLength(2);
    expect(resolved.graveyard[1].ID).toBe('a');
  });
});
