export type CommanderPick = {
  ID: string;
  Name: string;
  Text?: string | null;
};

function normalize(text?: string | null): string {
  return (text ?? '').toLowerCase();
}

export function hasPartnerAbility(card: CommanderPick): boolean {
  const text = normalize(card.Text);
  // Basic Partner and Partner with both include the word "partner".
  // Word boundary avoids matching unrelated substrings.
  return /\bpartner\b/.test(text);
}

export function partnerWithTargetName(card: CommanderPick): string | null {
  const text = card.Text ?? '';
  // Typical Oracle text looks like:
  // "Partner with X (When this creature enters the battlefield, target player ... )"
  // We capture up to the first '(' or newline.
  const match = text.match(/Partner with\s+([^\n(]+)/i);
  if (!match?.[1]) return null;
  const raw = match[1].trim();
  // Strip trailing punctuation that sometimes appears in odd text sources.
  return raw.replace(/[\s\.,;:]+$/, '').trim() || null;
}

export function canAddSecondCommander(selected: CommanderPick[]): boolean {
  if (selected.length === 0) return true;
  if (selected.length >= 2) return false;
  return hasPartnerAbility(selected[0]);
}

export function isValidPartnerPair(a: CommanderPick, b: CommanderPick): boolean {
  if (!hasPartnerAbility(a) || !hasPartnerAbility(b)) return false;

  const aTarget = partnerWithTargetName(a);
  if (aTarget && b.Name !== aTarget) return false;

  const bTarget = partnerWithTargetName(b);
  if (bTarget && a.Name !== bTarget) return false;

  return true;
}

export function partnerConstraintMessage(first: CommanderPick): string {
  const target = partnerWithTargetName(first);
  if (target) return `Second commander must be ${target}.`;
  if (!hasPartnerAbility(first)) return 'Selected commander does not have Partner.';
  return 'Second commander must have Partner.';
}
