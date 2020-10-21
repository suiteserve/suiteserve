export type Id = string | number;

export interface Entity {
  id: Id;
}

export interface VersionedEntity {
  version: number;
}

export interface SoftDeleteEntity {
  deleted?: boolean;
  deleted_at?: number;
}

export interface Attachment extends Entity, VersionedEntity, SoftDeleteEntity {
  suite_id?: Id;
  case_id?: Id;
  filename: string;
  content_type: string;
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
  planned_cases?: number;
  status: SuiteStatus | string;
  result?: SuiteResult | string;
  disconnected_at?: number;
  started_at: number;
  finished_at?: number;
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
  suite_id: Id;
  name?: string;
  description?: string;
  tags?: string[];
  idx: number;
  args?: {
    [key: string]: JsonValue;
  };
  status: CaseStatus | string;
  result?: CaseResult | string;
  created_at: number;
  started_at?: number;
  finished_at?: number;
}

export interface LogLine extends Entity {
  case_id: Id;
  idx: number;
  error?: boolean;
  line?: string;
}
