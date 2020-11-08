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
  readonly delete?: (keyof E)[];
}

export function applyWatchEvent<E extends Watchable>(
  evt: WatchEvent<E>,
): (es: E[]) => E[] | undefined {
  return (es) => {
    const e = es.find((e) => e.id === evt.id);
    if (e === undefined) {
      if (evt.insert === undefined) {
        return undefined;
      }
      return es.concat(evt.insert);
    }
    if (
      isVersionedEntity(evt.update) &&
      isVersionedEntity(e) &&
      evt.update.version < e.version
    ) {
      return es;
    }
    let res: E = {
      ...e,
    };
    if (evt.update !== undefined) {
      for (const k in evt.update) {
        // noinspection JSUnfilteredForInLoop
        res = {
          ...e,
          [k]: evt.update[k],
        };
      }
    }
    evt.delete?.forEach((k) => delete res[k]);
    return es.filter((e) => e.id !== evt.id).concat(res);
  };
}
