export type ActivationErrorEntry = Readonly<{
  title: string;
  body: string;
  affordance: string;
}>;

export const ACTIVATION_ERRORS = Object.freeze({
  provider_unavailable: Object.freeze({
    title: "We can't reach that deck link right now.",
    body: 'Paste your decklist instead — it takes about 10 seconds.',
    affordance: 'Paste instead',
  }),
  preview_error: Object.freeze({
    title: "We couldn't fully read a few cards.",
    body: 'Fix the highlighted entries below, then continue — your original paste is untouched.',
    affordance: 'Fix these cards',
  }),
  guest_session_error: Object.freeze({
    title: "We couldn't set up your seat at the table.",
    body: 'Nothing you entered was lost. Try again in a moment.',
    affordance: 'Try again',
  }),
  create_error: Object.freeze({
    title: "We couldn't create your table.",
    body: 'Your deck and commander picks are still here — try again.',
    affordance: 'Try again',
  }),
} satisfies Record<string, ActivationErrorEntry>);

export type ActivationErrorCode = keyof typeof ACTIVATION_ERRORS;

export const GENERIC_ACTIVATION_ERROR: ActivationErrorEntry = Object.freeze({
  title: "We couldn't finish that step.",
  body: 'Nothing you entered was lost. Please try again.',
  affordance: 'Try again',
});

function activationCode(error: unknown): string | undefined {
  if (typeof error !== 'object' || error === null || !('graphQLErrors' in error)) return undefined;
  const graphQLErrors = (error as { graphQLErrors?: unknown }).graphQLErrors;
  if (!Array.isArray(graphQLErrors) || graphQLErrors.length === 0) return undefined;
  const first = graphQLErrors[0];
  if (typeof first !== 'object' || first === null || !('extensions' in first)) return undefined;
  const extensions = (first as { extensions?: unknown }).extensions;
  if (typeof extensions !== 'object' || extensions === null || !('code' in extensions)) return undefined;
  const code = (extensions as { code?: unknown }).code;
  return typeof code === 'string' ? code : undefined;
}

/** Resolve only allowlisted product copy; the source error's message is never returned. */
export function resolveActivationError(error: unknown): ActivationErrorEntry {
  const code = activationCode(error);
  if (code && Object.prototype.hasOwnProperty.call(ACTIVATION_ERRORS, code)) {
    return ACTIVATION_ERRORS[code as ActivationErrorCode];
  }
  return GENERIC_ACTIVATION_ERROR;
}
