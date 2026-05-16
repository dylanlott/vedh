export interface TestUser {
  username: string;
  password: string;
}

export const CREATOR_DECKLIST = '100, Island';
export const JOINER_DECKLIST = '100, Swamp';

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
