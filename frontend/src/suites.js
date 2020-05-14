export async function fetchSuites(afterId, limit) {
  let url = new URL('/v1/suites', window.location.href);
  if (afterId) {
    url.searchParams.append('after_id', afterId);
  }
  if (limit) {
    url.searchParams.append('limit', limit);
  }

  const res = await fetch(url.href);
  const json = await res.json();

  if (res.ok) {
    return json;
  } else {
    throw `Error fetching suites: ${json.error}`;
  }
}
