export interface ZoneDefinition {
  ID: string;
  Label: string;
  Visibility: 'private' | 'public' | 'count_only';
  Kind: string;
  SupportsCards: boolean;
}

export interface GameFormat {
  ID: 'EDH' | 'GENERIC_DUEL';
  Name: string;
  StartingLife: number;
  DefaultDeckSize: number;
  CommanderEnabled: boolean;
  PhaseSequence: string[];
  Zones: ZoneDefinition[];
}

export const formatRegistry: Record<GameFormat['ID'], GameFormat> = {
  EDH: {
    ID: 'EDH',
    Name: 'Commander',
    StartingLife: 40,
    DefaultDeckSize: 99,
    CommanderEnabled: true,
    PhaseSequence: ['pregame', 'untap', 'upkeep', 'draw', 'main', 'combat', 'main2', 'end'],
    Zones: [
      { ID: 'commander', Label: 'Commander', Visibility: 'public', Kind: 'stacked', SupportsCards: true },
      { ID: 'library', Label: 'Library', Visibility: 'count_only', Kind: 'library', SupportsCards: true },
      { ID: 'hand', Label: 'Hand', Visibility: 'private', Kind: 'hand', SupportsCards: true },
      { ID: 'battlefield', Label: 'Battlefield', Visibility: 'public', Kind: 'grid', SupportsCards: true },
      { ID: 'graveyard', Label: 'Graveyard', Visibility: 'public', Kind: 'stacked', SupportsCards: true },
      { ID: 'exiled', Label: 'Exile', Visibility: 'public', Kind: 'stacked', SupportsCards: true },
      { ID: 'revealed', Label: 'Revealed', Visibility: 'public', Kind: 'stacked', SupportsCards: true },
      { ID: 'controlled', Label: 'Controlled', Visibility: 'public', Kind: 'grid', SupportsCards: true },
    ],
  },
  GENERIC_DUEL: {
    ID: 'GENERIC_DUEL',
    Name: 'Generic Duel',
    StartingLife: 20,
    DefaultDeckSize: 60,
    CommanderEnabled: false,
    PhaseSequence: ['draw', 'main', 'battle', 'end'],
    Zones: [
      { ID: 'deck', Label: 'Deck', Visibility: 'count_only', Kind: 'library', SupportsCards: true },
      { ID: 'hand', Label: 'Hand', Visibility: 'private', Kind: 'hand', SupportsCards: true },
      { ID: 'field', Label: 'Field', Visibility: 'public', Kind: 'grid', SupportsCards: true },
      { ID: 'discard', Label: 'Discard', Visibility: 'public', Kind: 'stacked', SupportsCards: true },
      { ID: 'banished', Label: 'Banished', Visibility: 'public', Kind: 'stacked', SupportsCards: true },
    ],
  },
};

export const formats = Object.values(formatRegistry);

export function lookupFormat(id?: string | null): GameFormat {
  if (id && id in formatRegistry) return formatRegistry[id as GameFormat['ID']];
  return formatRegistry.EDH;
}
