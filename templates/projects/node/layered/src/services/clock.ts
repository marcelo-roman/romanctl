export type Clock = () => Date;

export const systemClock: Clock = () => new Date();

export function isoDate(moment: Date): string {
  return moment.toISOString().slice(0, 10);
}
