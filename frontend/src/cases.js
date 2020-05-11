export async function fetchCases(suiteId) {
  const res = await fetch(`/v1/suites/${suiteId}/cases`);
  const json = await res.json();

  if (res.ok) {
    return json;
  } else {
    throw `Error fetching cases: ${json.error}`;
  }
}
