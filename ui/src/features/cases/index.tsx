import React from 'react';
import * as api from '../../api';
import { Link, NavLink, useParams } from 'react-router-dom';

export const Cases: React.FC = () => {
  const { suiteId } = useParams<{ suiteId: string }>();
  const cases = api.SAMPLE_CASES.filter(
    (c) => c.suite_id.toString() === suiteId
  );
  const attachments = api.SAMPLE_ATTACHMENTS.filter(
    (a) => a.suite_id?.toString() === suiteId
  );
  return (
    <div>
      <p>
        <Link to='/'>All Suites</Link>
      </p>
      <h1>Cases for Suite {suiteId}</h1>
      <ul>
        {cases.map((c) => (
          <li key={c.id}>
            <NavLink to={`/suites/${c.suite_id}/cases/${c.id}`}>
              {c.name}
            </NavLink>
            :&nbsp;
            <span
              style={{
                color:
                  c.status === api.CaseStatus.STARTED ? 'green' : undefined,
              }}
            >
              {c.status}
            </span>
            ,&nbsp;
            <span
              style={{
                color:
                  c.result === api.CaseResult.PASSED
                    ? 'green'
                    : c.result === api.CaseResult.SKIPPED ||
                      c.result === api.CaseResult.ABORTED
                    ? 'gray'
                    : undefined,
              }}
            >
              {c.result}
            </span>
            <br />
            <small>{new Date(c.created_at).toLocaleString()}</small>
          </li>
        ))}
      </ul>
      <h1>Attachments for Suite {suiteId}</h1>
      <ul>
        {attachments.map((a) => (
          <li key={a.id}>{a.filename}</li>
        ))}
      </ul>
    </div>
  );
};
