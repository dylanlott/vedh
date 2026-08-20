import type { CommanderPick } from './commanderPartner';

export const QUICK_START_DRAFT_KEY = 'edhgo/quickstart-draft';

export type QuickStartDraft = {
  deckText: string;
  sourceURL: string;
  corrections: Record<number, string>;
  selectedCommanders: CommanderPick[];
  displayName: string;
};

function getSessionStorage(): Storage | null {
  try {
    if (typeof sessionStorage !== 'undefined' && sessionStorage) return sessionStorage;
  } catch {
    // Ignore and try the window-scoped accessor.
  }
  try {
    if (typeof window !== 'undefined' && window.sessionStorage) return window.sessionStorage;
  } catch {
    // Storage is unavailable in this environment.
  }
  return null;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function isCommander(value: unknown): value is CommanderPick {
  return isRecord(value)
    && typeof value.ID === 'string'
    && typeof value.Name === 'string'
    && (value.Text === undefined || value.Text === null || typeof value.Text === 'string');
}

function isDraft(value: unknown): value is QuickStartDraft {
  if (!isRecord(value)) return false;
  if (
    typeof value.deckText !== 'string'
    || typeof value.sourceURL !== 'string'
    || typeof value.displayName !== 'string'
    || !isRecord(value.corrections)
    || !Array.isArray(value.selectedCommanders)
  ) return false;
  return Object.entries(value.corrections).every(([line, name]) => /^\d+$/.test(line) && typeof name === 'string')
    && value.selectedCommanders.every(isCommander);
}

export function readDraft(): QuickStartDraft | null {
  try {
    const storage = getSessionStorage();
    const raw = storage?.getItem(QUICK_START_DRAFT_KEY);
    if (!raw) return null;
    const parsed: unknown = JSON.parse(raw);
    if (!isDraft(parsed)) {
      storage?.removeItem(QUICK_START_DRAFT_KEY);
      return null;
    }
    return parsed;
  } catch {
    try {
      getSessionStorage()?.removeItem(QUICK_START_DRAFT_KEY);
    } catch {
      // A denied storage surface is already effectively absent.
    }
    return null;
  }
}

export function saveDraft(draft: QuickStartDraft): void {
  try {
    getSessionStorage()?.setItem(QUICK_START_DRAFT_KEY, JSON.stringify(draft));
  } catch {
    // A disabled storage surface must not block the activation funnel.
  }
}

export function clearDraft(): void {
  try {
    getSessionStorage()?.removeItem(QUICK_START_DRAFT_KEY);
  } catch {
    // A disabled storage surface is already effectively cleared.
  }
}
