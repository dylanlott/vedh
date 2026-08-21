export type DeckImportContext = Readonly<{
  gameLabel: string;
  formatLabel: string;
}>;

export const MAGIC_COMMANDER_DECK_CONTEXT = {
  gameLabel: 'Magic: The Gathering',
  formatLabel: 'Commander (EDH)',
} as const satisfies DeckImportContext;
