// Temporary hand-written GraphQL types until codegen is added.
export interface LoginMutation {
  login: {
    ID: string;
    Username: string;
    Token: string;
  };
}

export interface LoginMutationVariables {
  username: string;
  password: string;
}

export interface SignupMutation {
  signup: {
    ID: string;
    Username: string;
    Token: string;
  };
}

export interface SignupMutationVariables {
  username: string;
  password: string;
}

// --- Guest activation (Phase 2, plan 02-01) ---

export interface GuestSessionMutation {
  guestSession: {
    ID: string;
    Username: string;
    DisplayName?: string | null;
    IsGuest?: boolean | null;
    Token: string;
    GuestCredential?: string | null;
  };
}

export interface GuestSessionMutationVariables {
  displayName?: string | null;
  sessionID: string;
}

// --- Game creation (Phase 2, plan 02-04) ---

export interface InputCreateGame {
  ID: string;
  Handle?: string | null;
  FormatID?: string | null;
  SessionID?: string | null;
  Turn: {
    Player: string;
    Phase: string;
    Number: number;
    Priority: string;
  };
  Players: Array<Record<string, unknown>>;
}

export interface CreateGameMutationVariables {
  input: InputCreateGame;
}

// --- Deck import / product-event contract (Phase 1, plan 01-01) ---
// This file has no real codegen tool wired up yet (see the header comment
// above), so these mirror server/schema.graphql by hand. previewDeck and
// trackProductEvent are Mutation fields, not Query fields, per the locked
// api-contract in .planning/intel/constraints.md.

export interface Card {
  FaceName?: string | null;
  Name: string;
  ID: string;
  Colors?: string | null;
  ColorIdentity?: string | null;
  CMC?: string | null;
  ManaCost?: string | null;
  UUID?: string | null;
  Power?: string | null;
  Toughness?: string | null;
  Types?: string | null;
  Subtypes?: string | null;
  Supertypes?: string | null;
  Text?: string | null;
  TCGID?: string | null;
  ScryfallID?: string | null;
  SetCode?: string | null;
  CollectorNumber?: string | null;
  Category?: string | null;
  SourceFormat?: string | null;
}

export interface DeckSuggestion {
  Name: string;
  Score: number;
  LowConfidence: boolean;
}

export interface DeckImportIssue {
  SourceLine: number;
  RawLine: string;
  Name: string;
  Reason: string;
  Candidates: DeckSuggestion[];
}

export interface DeckPreviewEntry {
  Quantity: number;
  Name: string;
  SetCode?: string | null;
  CollectorNumber?: string | null;
  Category?: string | null;
  Section: string;
  SourceLine: number;
  Resolved: boolean;
  Card?: Card | null;
}

export interface DeckPreview {
  SourceType: string;
  CardCount: number;
  Entries: DeckPreviewEntry[];
  CommanderCandidates: Card[];
  Unresolved: DeckImportIssue[];
  Warnings: string[];
  CanContinue: boolean;
  // Additive extension of the locked DeckPreview block — see the dated
  // addendum in .planning/intel/constraints.md.
  BlockingErrors: string[];
}

export interface InputDeckImport {
  text?: string | null;
  sourceURL?: string | null;
  sessionID: string;
}

export interface PreviewDeckMutation {
  previewDeck: DeckPreview;
}

export interface PreviewDeckMutationVariables {
  input: InputDeckImport;
}

export interface InputProductEventMeta {
  key: string;
  value: string;
}

export interface InputProductEvent {
  name: string;
  sessionID: string;
  gameID?: string | null;
  role?: string | null;
  source?: string | null;
  outcome?: string | null;
  durationMs?: number | null;
  metadata?: InputProductEventMeta[] | null;
}

export interface TrackProductEventMutation {
  trackProductEvent: boolean;
}

export interface TrackProductEventMutationVariables {
  input: InputProductEvent;
}
