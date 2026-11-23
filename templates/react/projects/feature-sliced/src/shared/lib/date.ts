const dateTime = new Intl.DateTimeFormat('en-US', { dateStyle: 'short', timeStyle: 'short' });

export function formatDateTime(date: Date): string {
  return dateTime.format(date);
}
