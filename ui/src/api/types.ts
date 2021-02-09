import { EntityAdapter, EntityState } from '@reduxjs/toolkit';

export type Id = string;

export interface Entity {
  readonly id: Id;
}

export interface VersionedEntity extends Entity {
  readonly version: number;
}

function isVersionedEntity(e: any): e is VersionedEntity {
  return typeof e.version === 'number';
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
  readonly project?: string;
  readonly tags?: string[];
  readonly plannedCases?: number;
  readonly status: SuiteStatus;
  readonly result?: SuiteResult;
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

export interface Case extends Entity, VersionedEntity {
  readonly suiteId: Id;
  readonly name?: string;
  readonly description?: string;
  readonly tags?: string[];
  readonly idx: number;
  readonly status: CaseStatus;
  readonly result?: CaseResult;
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
  readonly insert?: E;
  readonly update?: Partial<E>;
}

export interface InsertWatchEvent<E extends Watchable> extends Entity {
  readonly insert: E;
}

export function isInsertWatchEvent<E extends Watchable>(
  evt: WatchEvent<E>
): evt is InsertWatchEvent<E> {
  return evt.insert !== undefined;
}

export interface UpdateWatchEvent<E extends Watchable> extends Entity {
  readonly update: Partial<E>;
}

export function isUpdateWatchEvent<E extends Watchable & VersionedEntity>(
  evt: WatchEvent<E>
): evt is UpdateWatchEvent<E> {
  return evt.update !== undefined;
}

export function onEntityInserted<E extends Entity>(
  adapter: EntityAdapter<E>,
  state: EntityState<E>,
  inserted: E
): EntityState<E> {
  const existing = adapter.getSelectors().selectById(state, inserted.id);
  if (existing !== undefined) {
    if (
      isVersionedEntity(existing) &&
      isVersionedEntity(inserted) &&
      inserted.version <= existing.version
    ) {
      return state;
    }
    adapter.removeOne(state, inserted.id);
  }
  return adapter.addOne(state, inserted);
}

export function onEntityUpdated<E extends Watchable & VersionedEntity>(
  adapter: EntityAdapter<E>,
  state: EntityState<E>,
  evt: UpdateWatchEvent<E>
): EntityState<E> {
  const existing = adapter.getSelectors().selectById(state, evt.id);
  if (existing === undefined) {
    throw new Error('Cannot update nonexistent entity.');
  }
  if (
    evt.update.version === undefined ||
    evt.update.version <= existing.version
  ) {
    return state;
  }
  return adapter.updateOne(state, {
    id: evt.id,
    changes: evt.update,
  });
}
