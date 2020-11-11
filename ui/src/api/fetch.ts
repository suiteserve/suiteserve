import * as api from './types';

async function fetchJSON<T>(url: string): Promise<T> {
  const resp = await fetch(url);
  if (!resp.ok) {
    throw new Error(resp.statusText);
  }
  return resp.json();
}

export async function fetchAllAttachments(): Promise<api.Attachment[]> {
  return fetchJSON('/v1/attachments');
}

export async function fetchAttachment(id: api.Id): Promise<api.Attachment> {
  return fetchJSON(`/v1/attachments/${encodeURIComponent(id)}`);
}

export async function fetchCase(id: api.Id): Promise<api.Case> {
  return fetchJSON(`/v1/cases/${encodeURIComponent(id)}`);
}

export async function fetchSuiteCases(suiteId: api.Id): Promise<api.Case[]> {
  return fetchJSON(`/v1/suites/${encodeURIComponent(suiteId)}/cases`);
}

export async function fetchCaseAttachments(
  id: api.Id
): Promise<api.Attachment[]> {
  return fetchJSON(`/v1/attachments?case=${encodeURIComponent(id)}`);
}

export async function fetchLogLine(id: api.Id): Promise<api.LogLine> {
  return fetchJSON(`/v1/logs/${encodeURIComponent(id)}`);
}

export async function fetchCaseLogs(caseId: api.Id): Promise<api.LogLine[]> {
  return fetchJSON(`/v1/cases/${encodeURIComponent(caseId)}/logs`);
}

export async function fetchSuite(id: api.Id): Promise<api.Suite> {
  return fetchJSON(`/v1/suites/${encodeURIComponent(id)}`);
}

export async function fetchSuiteAttachments(
  id: api.Id
): Promise<api.Attachment[]> {
  return fetchJSON(`/v1/attachments?suite=${encodeURIComponent(id)}`);
}

export async function fetchSuitePage(): Promise<api.SuitePage> {
  return fetchJSON('/v1/suites');
}

export async function fetchSuitePageAfter(
  cursor: api.SuitePageCursor
): Promise<api.SuitePage> {
  return fetchJSON(`/v1/suites?from=${encodeURIComponent(cursor)}`);
}
