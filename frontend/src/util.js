/**
 * Retries the given async function with some arguments until it succeeds.
 * Before awaiting a result and after getting a result, the continue function
 * is called to see if we still care about the result. If not, an exception is
 * thrown.
 * @param {function(): boolean} continueFn
 * @param {function(...): T} asyncFn
 * @param {...} args
 * @returns {Promise<T>}
 * @template T
 */
export async function retry(continueFn, asyncFn, ...args) {
  if (!continueFn()) throw 'Cancelled';
  try {
    const res = await asyncFn.bind(this)(...args);
    if (continueFn()) return res;
  } catch (err) {
    console.error(err);
    return await new Promise((resolve, reject) => setTimeout(
      () => retry.bind(this)(continueFn, asyncFn, ...args)
        .then(resolve)
        .catch(reject),
      5000));
  }
  throw 'Cancelled';
}

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
