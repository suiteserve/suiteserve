import * as api from './types';
import React from 'react';

export * from './types';

export interface Source {
  getAttachment(id: api.Id): Promise<api.Attachment>;

  getSuiteAttachments(id: api.Id): Promise<api.Attachment[]>;

  getCaseAttachments(id: api.Id): Promise<api.Attachment[]>;

  getAllAttachments(): Promise<api.Attachment[]>;

  getSuite(id: api.Id): Promise<api.Suite>;

  watch<E extends api.Watchable>(coll: string, handler: (evt: api.WatchEvent<E>) => void): void;

  unwatch(coll: string): void;

  getSuitePage(): Promise<api.SuitePage>;

  getSuitePageAfter(id: api.Id): Promise<api.SuitePage>;

  getCase(id: api.Id): Promise<api.Case>;

  getLogLine(id: api.Id): Promise<api.LogLine>;
}

export class ServerSource implements Source {
  private sse = new EventSource('/v1/suites?watch=true');
  private sseListeners: {
    [key: string]: EventListener;
  } = {};

  async getAllAttachments(): Promise<api.Attachment[]> {
    return getJSON('/v1/attachments');
  }

  async getAttachment(id: api.Id): Promise<api.Attachment> {
    return getJSON(`/v1/attachments/${encodeURIComponent(id)}`);
  }

  async getCase(id: api.Id): Promise<api.Case> {
    return getJSON(`/v1/cases/${encodeURIComponent(id)}`);
  }

  async getSuiteCases(suiteId: api.Id): Promise<api.Case[]> {
    return getJSON(`/v1/suites/${encodeURIComponent(suiteId)}/cases`);
  }

  async getCaseAttachments(id: api.Id): Promise<api.Attachment[]> {
    return getJSON(`/v1/attachments?case=${encodeURIComponent(id)}`);
  }

  async getLogLine(id: api.Id): Promise<api.LogLine> {
    return getJSON(`/v1/logs/${encodeURIComponent(id)}`);
  }

  async getCaseLogs(caseId: api.Id): Promise<api.LogLine[]> {
    return getJSON(`/v1/cases/${encodeURIComponent(caseId)}/logs`);
  }

  async getSuite(id: api.Id): Promise<api.Suite> {
    return getJSON(`/v1/suites/${encodeURIComponent(id)}`);
  }

  async getSuiteAttachments(id: api.Id): Promise<api.Attachment[]> {
    return getJSON(`/v1/attachments?suite=${encodeURIComponent(id)}`);
  }

  async getSuitePage(): Promise<api.SuitePage> {
    return getJSON('/v1/suites');
  }

  async getSuitePageAfter(cursor: api.SuitePageCursor): Promise<api.SuitePage> {
    return getJSON(`/v1/suites?from=${encodeURIComponent(cursor)}`);
  }

  public watch<E extends api.Watchable>(coll: string,
                                        handler: (evt: api.WatchEvent<E>) => void): void {
    this.sse.addEventListener(coll, this.sseListeners[coll] = ((evt: MessageEvent) => {
      handler(evt.data);
    }) as EventListener);
  }

  public unwatch(coll: string): void {
    this.sse.removeEventListener(coll, this.sseListeners[coll]);
    delete this.sseListeners[coll];
  }
}

async function getJSON<T>(url: string): Promise<T> {
  const resp = await fetch(url);
  if (!resp.ok) {
    throw new Error(resp.statusText);
  }
  return resp.json();
}

export const APIContext = React.createContext<Source>(new ServerSource());
