export async function fetchSuites() {
  const res = await fetch('/suites');
  const json = await res.json();

  if (res.ok) {
    return json;
  } else {
    throw `Error fetching suites: ${json.error}`;
  }
}
