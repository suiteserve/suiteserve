export async function fetchSuites(fromId, limit) {
  const url = new URL('/v1/suites', window.location.href);
  if (fromId) {
    url.searchParams.append('from_id', fromId);
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
