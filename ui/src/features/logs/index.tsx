import React from 'react';
import * as api from '../../api';
import { LogLevelType } from '../../api';
import { Link, useParams } from 'react-router-dom';

export const Logs: React.FC = () => {
  const { suiteId, caseId } = useParams<{ suiteId: string; caseId: string }>();
  const logs = api.SAMPLE_LOGS.filter(
    (l) => l.case_id.toString() === caseId
  ).sort((a, b) => {
    if (a.timestamp === b.timestamp) {
      return b.idx - a.idx;
    }
    return b.timestamp - a.timestamp;
  });
  const attachments = api.SAMPLE_ATTACHMENTS.filter(
    (a) => a.case_id?.toString() === caseId
  );
  return (
    <div>
      <p>
        <Link to={`/`}>All Suites</Link> /{' '}
        <Link to={`/suites/${suiteId}`}>{suiteId}</Link>
      </p>
      <h1>
        Logs for Case {caseId} in Suite {suiteId}
      </h1>
      <table>
        <tbody>
          {logs.map((l) => (
            <tr
              key={l.id}
              style={{
                fontFamily:
                  'SFMono-Regular,Consolas,Liberation Mono,Menlo,monospace',
                fontSize: '12px',
              }}
            >
              <td
                style={{
                  padding: '0.25em 0.5em',
                }}
              >
                {new Date(l.timestamp).toISOString()}
              </td>
              <td
                style={{
                  padding: '0.25em 0.5em',
                  color: l.level === LogLevelType.ERROR ? 'red' : undefined,
                }}
              >
                {l.level}
              </td>
              <td
                style={{
                  padding: '0.25em 0.5em',
                }}
              >
                {l.message}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      <h1>Attachments for Case {caseId}</h1>
      <ul>
        {attachments.map((a) => (
          <li key={a.id}>{a.filename}</li>
        ))}
      </ul>
    </div>
  );
};
