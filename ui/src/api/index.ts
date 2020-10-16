import * as api from './types';

export * from './types';

export interface Source {

  getAttachment(id: api.Id): Promise<api.Attachment>;

  getSuiteAttachments(id: api.Id): Promise<api.Attachment[]>;

  getCaseAttachments(id: api.Id): Promise<api.Attachment[]>;

  getAllAttachments(): Promise<api.Attachment[]>;

  getSuite(id: api.Id): Promise<api.Suite>;

  watch(suiteId: api.Id): EventSource;

  getSuitePage(): Promise<api.SuitePage>;

  getSuitePageAfter(id: api.Id): Promise<api.SuitePage>;

  getCase(id: api.Id): Promise<api.Case>;

  getLogLine(id: api.Id): Promise<api.LogLine>;
}

export class ServerSource implements Source {
  async getAllAttachments(): Promise<api.Attachment[]> {
    return getJson('/v1/attachments');
  }

  async getAttachment(id: api.Id): Promise<api.Attachment> {
    return getJson(`/v1/attachments/${encodeURIComponent(id)}`);
  }

  async getCase(id: api.Id): Promise<api.Case> {
    return getJson(`/v1/cases/${encodeURIComponent(id)}`);
  }

  async getSuiteCases(suiteId: api.Id): Promise<api.Case[]> {
    return getJson(`/v1/suites/${encodeURIComponent(suiteId)}/cases`);
  }

  async getCaseAttachments(id: api.Id): Promise<api.Attachment[]> {
    return getJson(`/v1/attachments?case=${encodeURIComponent(id)}`);
  }

  async getLogLine(id: api.Id): Promise<api.LogLine> {
    return getJson(`/v1/logs/${encodeURIComponent(id)}`);
  }

  async getCaseLogs(caseId: api.Id): Promise<api.LogLine[]> {
    return getJson(`/v1/cases/${encodeURIComponent(caseId)}/logs`);
  }

  async getSuite(id: api.Id): Promise<api.Suite> {
    return getJson(`/v1/suites/${encodeURIComponent(id)}`);
  }

  async getSuiteAttachments(id: api.Id): Promise<api.Attachment[]> {
    return getJson(`/v1/attachments?suite=${encodeURIComponent(id)}`);
  }

  async getSuitePage(): Promise<api.SuitePage> {
    return getJson('/v1/suites');
  }

  async getSuitePageAfter(id: api.Id): Promise<api.SuitePage> {
    return getJson(`/v1/suites?after=${encodeURIComponent(id)}`);
  }

  public watch(suiteId: api.Id): EventSource {
    return new EventSource('/v1/suites?watch=true');
  }
}

async function getJson<T>(url: string): Promise<T> {
  const resp = await fetch(url);
  if (!resp.ok) {
    throw new Error(resp.statusText);
  }
  return resp.json();
}
