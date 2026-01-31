export type StackCard = {
  ID?: string;
  Name: string;
  Types?: string;
  CurrentZone?: string;
  Tapped?: boolean;
};

export type MoveHandToStackResult = {
  hand: StackCard[];
  stack: StackCard[];
  movedCard?: StackCard;
  skippedReason?: 'not_found' | 'duplicate';
};

export function isLandCard(card: StackCard | null | undefined): boolean {
  const types = card?.Types;
  return typeof types === 'string' && types.includes('Land');
}

export function moveHandCardToStackState(
  hand: StackCard[],
  stack: StackCard[],
  card: StackCard,
  owner: string,
  fromIndex?: number,
): MoveHandToStackResult {
  const nextHand = [...hand];
  const nextStack = [...stack];
  const id = card?.ID;
  if (id && nextStack.some(c => c.ID === id)) {
    return { hand: nextHand, stack: nextStack, skippedReason: 'duplicate' };
  }

  let removeIndex = -1;
  if (id) {
    removeIndex = nextHand.findIndex(c => c.ID === id);
  }
  if (removeIndex === -1 && typeof fromIndex === 'number') {
    removeIndex = fromIndex;
  }
  if (removeIndex < 0 || removeIndex >= nextHand.length) {
    return { hand: nextHand, stack: nextStack, skippedReason: 'not_found' };
  }

  const moved = { ...nextHand[removeIndex], CurrentZone: owner, Tapped: false };
  const updatedHand = nextHand.filter((_, idx) => idx !== removeIndex);
  nextStack.push(moved);

  return { hand: updatedHand, stack: nextStack, movedCard: moved };
}

export type ResolveStackResult = {
  stack: StackCard[];
  graveyard: StackCard[];
  movedCard?: StackCard;
  skippedReason?: 'not_found';
};

export function resolveStackCardToGraveyardState(
  stack: StackCard[],
  graveyard: StackCard[],
  stackIndex: number,
): ResolveStackResult {
  if (stackIndex < 0 || stackIndex >= stack.length) {
    return { stack: [...stack], graveyard: [...graveyard], skippedReason: 'not_found' };
  }
  const nextStack = [...stack];
  const [moved] = nextStack.splice(stackIndex, 1);
  const nextGraveyard = [...graveyard, moved];
  return { stack: nextStack, graveyard: nextGraveyard, movedCard: moved };
}
