import * as api from './types';
import suites from './sample_suites.json';
import cases from './sample_cases.json';
import logs from './sample_logs.json';
import attachments from './sample_attachments.json';

export * from './types';

export const SAMPLE_SUITES: api.Suite[] = suites;
export const SAMPLE_CASES: api.Case[] = cases;
export const SAMPLE_LOGS: api.LogLine[] = logs;
export const SAMPLE_ATTACHMENTS: api.Attachment[] = attachments;
