import * as api from './types';

async function getJSON<T>(url: string): Promise<T> {
  const resp = await fetch(url);
  if (!resp.ok) {
    throw new Error(resp.statusText);
  }
  return resp.json();
}

export async function getAllAttachments(): Promise<api.Attachment[]> {
  return getJSON('/v1/attachments');
}

export async function getAttachment(id: api.Id): Promise<api.Attachment> {
  return getJSON(`/v1/attachments/${encodeURIComponent(id)}`);
}

export async function getCase(id: api.Id): Promise<api.Case> {
  return getJSON(`/v1/cases/${encodeURIComponent(id)}`);
}

export async function getSuiteCases(suiteId: api.Id): Promise<api.Case[]> {
  return getJSON(`/v1/suites/${encodeURIComponent(suiteId)}/cases`);
}

export async function getCaseAttachments(
  id: api.Id,
): Promise<api.Attachment[]> {
  return getJSON(`/v1/attachments?case=${encodeURIComponent(id)}`);
}

export async function getLogLine(id: api.Id): Promise<api.LogLine> {
  return getJSON(`/v1/logs/${encodeURIComponent(id)}`);
}

export async function getCaseLogs(caseId: api.Id): Promise<api.LogLine[]> {
  return getJSON(`/v1/cases/${encodeURIComponent(caseId)}/logs`);
}

export async function getSuite(id: api.Id): Promise<api.Suite> {
  return getJSON(`/v1/suites/${encodeURIComponent(id)}`);
}

export async function getSuiteAttachments(
  id: api.Id,
): Promise<api.Attachment[]> {
  return getJSON(`/v1/attachments?suite=${encodeURIComponent(id)}`);
}

export async function getSuitePage(): Promise<api.SuitePage> {
  return getJSON('/v1/suites');
}

export async function getSuitePageAfter(
  cursor: api.SuitePageCursor,
): Promise<api.SuitePage> {
  return getJSON(`/v1/suites?from=${encodeURIComponent(cursor)}`);
}
