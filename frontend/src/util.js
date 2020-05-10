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

export function formatTime (millis) {
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
