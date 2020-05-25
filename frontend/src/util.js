/**
 * Formats the given number of milliseconds since the Unix Epoch as a date in
 * the user's locale.
 * @param {int} millis
 * @returns {string}
 */
export function formatUnix(millis) {
  return new Date(millis).toLocaleString(navigator.languages, {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}
