/**
 * Formats the given number of milliseconds since the Unix Epoch as a date in
 * the user's locale.
 * @param {int} millis
 * @returns {string}
 */
export function formatTime(millis) {
  const date = new Date(millis);
  const opts = {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  };
  return date.toLocaleString(navigator.languages, opts);
}
