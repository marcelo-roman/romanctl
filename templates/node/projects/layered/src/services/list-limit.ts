export const DEFAULT_LIST_LIMIT = 50;

export const MAX_LIST_LIMIT = 200;

export function listLimit(requested: number | undefined): number {
  if (requested === undefined || !Number.isInteger(requested)) {
    return DEFAULT_LIST_LIMIT;
  }
  return Math.min(Math.max(requested, 1), MAX_LIST_LIMIT);
}
