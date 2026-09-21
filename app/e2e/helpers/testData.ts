export interface TestUser {
  username: string;
  password: string;
}

export const CREATOR_DECKLIST = "Commander\n1 Kykar, Wind's Fury\nDeck\n99 Island";
export const JOINER_DECKLIST = 'Commander\n1 Jarad, Golgari Lich Lord\nDeck\n99 Swamp';
export const E2E_RUN_ID = process.env.VEDH_E2E_RUN_ID || `local-${Date.now()}-${process.pid}`;

export function e2eSession(label: string): string {
  return `e2e-${E2E_RUN_ID}-${label}`;
}

function uniqueSuffix() {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

export function createTestUser(prefix: string): TestUser {
  const suffix = uniqueSuffix();

  return {
    username: `${prefix}-${suffix}`,
    password: `Pass!${suffix}`,
  };
}
