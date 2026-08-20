/**
 * The single rendering seam for player names.
 *
 * Username remains the unique identity and join key everywhere. Callers must
 * never compare, key, or authorize with this function's display-only output.
 */
export interface PlayerNameRecord {
  DisplayName?: string | null;
  Username?: string | null;
}

export function displayNameOf(player?: PlayerNameRecord | null): string {
  const displayName = player?.DisplayName;
  if (typeof displayName === 'string' && displayName.trim() !== '') {
    return displayName;
  }
  return typeof player?.Username === 'string' ? player.Username : '';
}
