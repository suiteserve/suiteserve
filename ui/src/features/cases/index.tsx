import React, { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import styles from './Cases.module.css';
import * as api from '../../api';

export const Cases: React.FC = () => {
  const { suiteId } = useParams<{ suiteId: string }>();
  const [cases, setCases] = useState([] as api.Case[]);

  useEffect(() => {
    new api.ServerSource().getSuiteCases(suiteId).then((cases) => {
      setCases(
        cases.sort((a, b) => {
          if (a.idx === b.idx) {
            return b.idx - a.idx;
          }
          return b.createdAt - a.createdAt;
        })
      );
    });
  }, [suiteId]);

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
              <td
                className={
                  c.status === api.CaseStatus.CREATED ? styles.Warn : ''
                }
              >
                {c.status}
              </td>
              <td
                className={
                  c.result === api.CaseResult.PASSED
                    ? styles.Good
                    : c.result === api.CaseResult.SKIPPED
                    ? styles.Warn
                    : styles.Bad
                }
              >
                {c.result}
              </td>
              <td>{new Date(c.createdAt).toISOString()}</td>
              <td>{!c.startedAt ? '' : new Date(c.startedAt).toISOString()}</td>
              <td>
                {!c.finishedAt ? '' : new Date(c.finishedAt).toISOString()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
