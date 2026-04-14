export type SuccessResultLike = {
  success?: boolean;
  redirect?: unknown;
};

export function getRedirectUrl(result: SuccessResultLike | undefined | null): string | null {
  if (!result?.success) return null;
  if (typeof result.redirect !== 'string') return null;
  const trimmed = result.redirect.trim();
  return trimmed.length > 0 ? trimmed : null;
}

