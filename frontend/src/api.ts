export interface Entity {
  readonly id: string;
}

export interface VersionedEntity extends Entity {
  readonly version: number;
}

export interface SoftDeleteEntity extends Entity {
  readonly deleted: boolean;
  readonly deleted_at?: number;
}

export interface Attachment extends VersionedEntity, SoftDeleteEntity {
  readonly size: number;
}

export interface SuiteFailureType {
  readonly name: string;
  readonly description?: string;
}

export interface SuiteEnvVar {
  readonly key: string;
  readonly value?: any;
}

export enum SuiteStatus {
  Running = 'running',
  Passed = 'passed',
  Failed = 'failed',
  Disconnected = 'disconnected',
}

export interface Suite extends VersionedEntity, SoftDeleteEntity {
  readonly name?: string;
  readonly failure_types?: ReadonlyArray<SuiteFailureType>;
  readonly tags?: ReadonlyArray<string>;
  readonly env_vars?: ReadonlyArray<SuiteEnvVar>;
  readonly attachments?: ReadonlyArray<string>;
  readonly planned_cases: number;
  readonly status: SuiteStatus;
  readonly started_at: number;
  readonly finished_at?: number;
  readonly disconnected_at?: number;
}

export interface SuiteAggs extends VersionedEntity {
  readonly id: string;
  readonly running: number;
  readonly finished: number;
}

export interface SuitePage {
  readonly aggs: SuiteAggs;
  readonly next_id?: string;
  readonly suites?: ReadonlyArray<Suite>;
}

export enum ErrorType {
  BadRequest = 'bad_request',
  NotFound = 'not_found',
  Unknown = 'unknown',
}

export interface Error {
  readonly error: ErrorType;
}

export enum CaseLinkType {
  Issue = 'issue',
  Other = 'other',
}

export enum CaseStatus {
  Created = 'created',
  Disabled = 'disabled',
  Running = 'running',
  Passed = 'passed',
  Failed = 'failed',
  Errored = 'errored',
  Aborted = 'aborted',
}

export interface CaseLink {
  readonly type: CaseLinkType;
  readonly name: string;
  readonly url: string;
}

export interface CaseArg {
  readonly key: string;
  readonly value?: any;
}

export interface Case extends VersionedEntity {
  readonly suite: string;
  readonly name: string;
  readonly description?: string;
  readonly tags?: ReadonlyArray<string>;
  readonly num: number;
  readonly links?: ReadonlyArray<CaseLink>;
  readonly args?: ReadonlyArray<CaseArg>;
  readonly attachments?: ReadonlyArray<string>;
  readonly status: CaseStatus;
  readonly created_at: number;
  readonly started_at?: number;
  readonly finished_at?: number;
}

export enum LogLevelType {
  Trace = 'trace',
  Debug = 'debug',
  Info = 'info',
  Warn = 'warn',
  Error = 'error',
}

export interface LogLine extends Entity {
  readonly case: string;
  readonly index: number;
  readonly level: LogLevelType;
  readonly trace?: string;
  readonly message?: string;
  readonly timestamp: number;
}

export interface LogPage {
  readonly next_id?: string;
  readonly lines?: ReadonlyArray<LogLine>;
}

export enum Coll {
  Attachments = 'attachments',
  Cases = 'cases',
  Logs = 'logs',
  Suites = 'suites',
  SuiteAggs = 'suite_aggs',
}

export enum ChangeMessageCmd {
  Change = 'change',
  Ok = 'ok',
  SubSuites = 'sub_suites',
}

export type ChangeMessagePayload<T extends ChangeMessageCmd> =
    T extends ChangeMessageCmd.Change ? Change :
        T extends ChangeMessageCmd.SubSuites ? SubSuitesChangeMessage :
            undefined

export interface ChangeMessage<Cmd extends ChangeMessageCmd = ChangeMessageCmd> {
  readonly seq: number;
  readonly cmd: Cmd;
  readonly payload: ChangeMessagePayload<Cmd>;
}

export function isChangeMessage(msg: ChangeMessage): msg is ChangeMessage<ChangeMessageCmd.Change> {
  return msg.cmd === ChangeMessageCmd.Change;
}

export function isOkMessage(msg: ChangeMessage): msg is ChangeMessage<ChangeMessageCmd.Ok> {
  return msg.cmd === ChangeMessageCmd.Ok;
}

export function isSubSuitesMessage(msg: ChangeMessage): msg is ChangeMessage<ChangeMessageCmd.SubSuites> {
  return msg.cmd === ChangeMessageCmd.SubSuites;
}

export interface SubSuitesChangeMessage {
  readonly ids: ReadonlyArray<string>;
}

export enum ChangeOp {
  Insert = 'insert',
  Update = 'update',
}

export type EntityColl<T extends Entity> =
    T extends Attachment ? Coll.Attachments :
        T extends Suite ? Coll.Suites :
            T extends Case ? Coll.Cases :
                T extends LogLine ? Coll.Logs :
                    Coll

export interface Change<Op extends ChangeOp = ChangeOp, T extends Entity = Entity> {
  readonly op: Op;
  readonly coll: EntityColl<T>;
  readonly id: string;
}

export function isSuiteInsertChange(change: Change): change is InsertChange<Suite> {
  return change.op === ChangeOp.Insert && change.coll === Coll.Suites;
}

export function isSuiteUpdateChange(change: Change): change is UpdateChange<Suite> {
  return change.op === ChangeOp.Update && change.coll === Coll.Suites;
}

export function isSuiteAggsInsertChange(change: Change): change is InsertChange<SuiteAggs> {
  return change.op === ChangeOp.Insert && change.coll === Coll.SuiteAggs;
}

export function isSuiteAggsUpdateChange(change: Change): change is UpdateChange<SuiteAggs> {
  return change.op === ChangeOp.Update && change.coll === Coll.SuiteAggs;
}

export interface InsertChange<T extends Entity = Entity>
    extends Change<ChangeOp.Insert, T> {
  readonly doc: T;
}

export interface UpdateChange<T extends VersionedEntity = VersionedEntity>
    extends Change<ChangeOp.Update, T> {
  readonly updated: Partial<Omit<T, 'version'>> & Pick<T, 'version'>;
  readonly deleted: ReadonlyArray<keyof Omit<T, 'version'>>;
}

export function applyUpdateChange<T extends VersionedEntity>(
    entity: T, update?: UpdateChange<T>): T {
  if (!update || entity.version >= update.updated.version) {
    return entity;
  }
  const newUpdate = {...entity, ...update.updated};
  update.deleted.forEach(field => delete newUpdate[field]);
  return newUpdate;
}

export function mergeUpdateChanges<T extends VersionedEntity>(
    a: UpdateChange<T> | undefined, b: UpdateChange<T>): UpdateChange<T> {
  if (!a) {
    return b;
  }
  const [oldest, newest] = [a, b].sort((a, b) =>
      a.updated.version - b.updated.version);
  return {
    ...oldest,
    updated: {...oldest.updated, ...newest.updated},
    deleted: [...oldest.deleted, ...newest.deleted],
  };
}

type FetchResponse<T extends {}> = T | Error;

function isErrorResponse<T extends {}>(res: FetchResponse<T>):
    res is Error {
  return res.hasOwnProperty('error');
}

export async function fetchSuitePage(options: { fromId?: string, limit?: number }):
    Promise<SuitePage> {
  const url = new URL('/v1/suites', window.location.href);
  if (options.fromId) {
    url.searchParams.append('from_id', options.fromId);
  }
  if (options.limit) {
    url.searchParams.append('limit', options.limit.toString());
  }

  const res = await fetch(url.href);
  const json: FetchResponse<SuitePage> = await res.json();

  if (isErrorResponse(json)) {
    throw `Fetch suite page: ${json.error}`;
  }
  return json;
}

export async function fetchCases(suiteId: string): Promise<ReadonlyArray<Case>> {
  const url = new URL(`/v1/suites/${suiteId}/cases`, window.location.href);
  const res = await fetch(url.href);
  const json: FetchResponse<ReadonlyArray<Case>> = await res.json();

  if (isErrorResponse(json)) {
    throw `Fetch cases by suite: ${json.error}`;
  }
  return json;
}
