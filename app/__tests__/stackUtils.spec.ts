import { describe, it, expect } from 'vitest';
import { isLandCard, moveHandCardToStackState, resolveStackCardToGraveyardState } from '../src/utils/stack';

describe('stack utils', () => {
  it('detects land cards by type', () => {
    expect(isLandCard({ Name: 'Forest', Types: 'Basic Land — Forest' })).toBe(true);
    expect(isLandCard({ Name: 'Shock', Types: 'Instant' })).toBe(false);
    expect(isLandCard(undefined)).toBe(false);
  });

  it('moves duplicate printings to the stack as separate physical cards', () => {
    const hand = [
      { ID: 'a', Name: 'Shock', Types: 'Instant' },
      { ID: 'a', Name: 'Shock', Types: 'Instant' },
    ];
    const stack: typeof hand = [];
    const moved = moveHandCardToStackState(hand, stack, hand[0], 'alice', 0);
    expect(moved.hand).toHaveLength(1);
    expect(moved.stack).toHaveLength(1);
    expect(moved.movedCard?.CurrentZone).toBe('alice');
    expect(moved.movedCard?.Tapped).toBe(false);

    const second = moveHandCardToStackState(moved.hand, moved.stack, moved.hand[0], 'alice', 0);
    expect(second.hand).toHaveLength(0);
    expect(second.stack).toHaveLength(2);
    expect(second.stack.map(card => card.ID)).toEqual(['a', 'a']);
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
