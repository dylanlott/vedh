import { describe, expect, it } from 'vitest';
import { commanderPickFor, deckCardCount, decklistToCsv, defaultDecks } from '../src/decks/defaultDecks';
describe('default decks', () => {
  it('provides one current, complete Commander list per bracket', () => {
    expect(defaultDecks.map(deck => deck.bracket)).toEqual([1, 2, 3, 4, 5]);
    expect(new Set(defaultDecks.map(deck => deck.id)).size).toBe(defaultDecks.length);
    for (const deck of defaultDecks) {
      expect(deckCardCount(deck), deck.id).toBe(100);
      expect(deck.updatedAt).toMatch(/^\d{4}-\d{2}-\d{2}$/);
      expect(deck.entries).toContainEqual([1, deck.commander]);
    }
  });
  it('serializes quoted names into backend-compatible CSV', () => {
    expect(decklistToCsv(defaultDecks[0])).toContain('1,"Pantlaza, Sun-Favored"');
    expect(commanderPickFor(defaultDecks[0]).Name).toBe('Pantlaza, Sun-Favored');
  });
});
