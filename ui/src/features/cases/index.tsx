import React from 'react';
import * as api from '../../api';
import { Link, NavLink, useParams } from 'react-router-dom';
import styles from './Cases.module.css';
import {CaseResult, CaseStatus} from '../../api';

export const Cases: React.FC = () => {
  const { suiteId } = useParams<{ suiteId: string }>();
  const cases = api.SAMPLE_CASES.filter(
    (c) => c.suite_id.toString() === suiteId
  ).sort((a, b) => {
    if (a.idx === b.idx) {
      return b.created_at - a.created_at;
    }
    return b.idx - a.idx;
  });
  const attachments = api.SAMPLE_ATTACHMENTS.filter(
    (a) => a.suite_id?.toString() === suiteId
  );
  return (
    <div className={styles.Cases}>
      <Link to='/'>All Suites</Link>
      <h1>Cases for Suite {suiteId}</h1>
      <table>
        <thead>
          <tr>
            <td>Index</td>
            <td>Name</td>
            <td>Tags</td>
            <td>Description</td>
            <td>Status</td>
            <td>Result</td>
            <td>Created At</td>
            <td>Started At</td>
            <td>Finished At</td>
          </tr>
        </thead>
        <tbody>
          {cases.map((c) => (
            <tr key={c.id}>
              <td>{c.idx}</td>
              <td>
                <Link to={`/suites/${suiteId}/cases/${c.id}`}>
                  {c.name || c.id}
                </Link>
              </td>
              <td>{c.tags?.join(', ')}</td>
              <td>{c.description}</td>
              <td className={c.status === CaseStatus.CREATED ? styles.Warn : ''}>{c.status}</td>
              <td className={c.result === CaseResult.PASSED ? styles.Good : styles.Bad}>{c.result}</td>
              <td>{new Date(c.created_at).toISOString()}</td>
              <td>
                {!c.started_at
                  ? ''
                  : new Date(c.started_at).toISOString()}
              </td>
              <td>
                {!c.finished_at
                  ? ''
                  : new Date(c.finished_at).toISOString()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
