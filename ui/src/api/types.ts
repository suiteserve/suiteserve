export type Id = string | number;

export interface Entity {
  id: Id;
}

export interface VersionedEntity {
  version: number;
}

export interface SoftDeleteEntity {
  deleted?: boolean;
  deletedAt?: number;
}

export interface Attachment extends Entity, VersionedEntity, SoftDeleteEntity {
  suiteId?: Id;
  caseId?: Id;
  filename: string;
  contentType: string;
  size: number;
  timestamp: number;
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

export interface Suite extends Entity, VersionedEntity, SoftDeleteEntity {
  name?: string;
  tags?: string[];
  plannedCases?: number;
  status: SuiteStatus | string;
  result?: SuiteResult | string;
  disconnectedAt?: number;
  startedAt: number;
  finishedAt?: number;
}

export interface SuitePage {
  more: boolean;
  suites: Suite[];
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
  suiteId: Id;
  name?: string;
  description?: string;
  tags?: string[];
  idx: number;
  args?: {
    [key: string]: JsonValue;
  };
  status: CaseStatus | string;
  result?: CaseResult | string;
  createdAt: number;
  startedAt?: number;
  finishedAt?: number;
}

export interface LogLine extends Entity {
  caseId: Id;
  idx: number;
  error?: boolean;
  line?: string;
}
