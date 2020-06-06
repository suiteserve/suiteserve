export async function fetchCases(suiteId) {
  const url = new URL(`/v1/suites/${suiteId}/cases`, window.location.href);
  const res = await fetch(url.href);
  const json = await res.json();

  if (res.ok) {
    return json;
  } else {
    throw `Error fetching cases: ${json.error}`;
  }
}
