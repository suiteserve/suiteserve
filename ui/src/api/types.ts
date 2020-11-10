import {EntityAdapter, EntityState} from '@reduxjs/toolkit';

export type Id = string;

export interface Entity {
  readonly id: Id;
}

export interface VersionedEntity extends Entity {
  readonly version: number;
}

function isVersionedEntity(e: any): e is VersionedEntity {
  return 'version' in e && typeof e['version'] === 'number';
}

export interface Attachment extends Entity, VersionedEntity {
  readonly suiteId?: Id;
  readonly caseId?: Id;
  readonly filename: string;
  readonly contentType: string;
  readonly size: number;
  readonly timestamp: number;
}

export enum SuiteStatus {
  STARTED = 'started',
  FINISHED = 'finished',
  DISCONNECTED = 'disconnected',
}

export enum SuiteResult {
  PASSED = 'passed',
  FAILED = 'failed',
}

export interface Suite extends Entity, VersionedEntity {
  readonly name?: string;
  readonly tags?: string[];
  readonly plannedCases?: number;
  readonly status: SuiteStatus | string;
  readonly result?: SuiteResult | string;
  readonly disconnectedAt?: number;
  readonly startedAt: number;
  readonly finishedAt?: number;
}

export type SuitePageCursor = string;

export interface SuitePage {
  readonly next: SuitePageCursor;
  readonly suites: Suite[];
}

export enum CaseStatus {
  CREATED = 'created',
  STARTED = 'started',
  FINISHED = 'finished',
}

export enum CaseResult {
  PASSED = 'passed',
  FAILED = 'failed',
  SKIPPED = 'skipped',
  ABORTED = 'aborted',
  ERRORED = 'errored',
}

type JsonValue =
  | string
  | number
  | boolean
  | null
  | Map<string, JsonValue>
  | Array<JsonValue>;

export interface Case extends Entity, VersionedEntity {
  readonly suiteId: Id;
  readonly name?: string;
  readonly description?: string;
  readonly tags?: string[];
  readonly idx: number;
  readonly args?: {
    [key: string]: JsonValue;
  };
  readonly status: CaseStatus | string;
  readonly result?: CaseResult | string;
  readonly createdAt: number;
  readonly startedAt?: number;
  readonly finishedAt?: number;
}

export interface LogLine extends Entity {
  readonly caseId: Id;
  readonly idx: number;
  readonly error?: boolean;
  readonly line?: string;
}

export type Watchable = Suite | Case | LogLine;

export interface WatchEvent<E extends Watchable> extends Entity {
  readonly type: string;
  readonly insert?: E;
  readonly update?: Partial<E>;
}

export function upsertEntity<E extends Entity>(
  adapter: EntityAdapter<E>,
  state: EntityState<E>,
  e: E,
): EntityState<E> {
  const existing = adapter.getSelectors().selectById(state, e.id);
  if (existing === undefined) {
    return adapter.addOne(state, e);
  } else if (
    isVersionedEntity(existing) &&
    isVersionedEntity(e) &&
    existing.version < e.version
  ) {
    return adapter.setAll(state, [e]);
  }
  return state;
}

export const upsertEntities = <E extends Entity>(
  adapter: EntityAdapter<E>,
  state: EntityState<E>,
  es: E[],
): EntityState<E> =>
  es.reduce((state, e) => upsertEntity(adapter, state, e), state);

export function applyWatchEvent<E extends Watchable>(
  adapter: EntityAdapter<E>,
  state: EntityState<E>,
  evt: WatchEvent<E>,
) {
  const existing = adapter.getSelectors().selectById(state, evt.id);
  if (existing === undefined) {
    if (evt.insert === undefined) {
      throw new Error('WatchEvent did not have a required insert field');
    } else {
      adapter.addOne(state, evt.insert);
    }
  } else if (
    isVersionedEntity(existing) &&
    isVersionedEntity(evt.update) &&
    existing.version >= evt.update.version
  ) {
    // do nothing
  } else if (evt.update !== undefined) {
    adapter.updateOne(state, {
      id: evt.id,
      changes: evt.update,
    });
  } else {
    throw new Error('WatchEvent did not have a required update field');
  }
}
