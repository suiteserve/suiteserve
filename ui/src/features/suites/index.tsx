import React from 'react';
import * as api from '../../api';
import {SuiteResult, SuiteStatus} from '../../api';
import styles from './Suites.module.css';
import {Link} from 'react-router-dom';

export const Suites: React.FC = () => {
  const suites = api.SAMPLE_SUITES.sort((a, b) => {
    return b.started_at - a.started_at;
  });
  return (
    <div className={styles.Suites}>
      <h1>All Suites</h1>
      <table>
        <thead>
          <tr>
            <td>Name</td>
            <td>Tags</td>
            <td>Planned Cases</td>
            <td>Status</td>
            <td>Result</td>
            <td>Started At</td>
            <td>Finished At</td>
            <td>Disconnected At</td>
          </tr>
        </thead>
        <tbody>
          {suites.map((suite) => (
            <tr key={suite.id}>
              <td>
                <Link to={`/suites/${suite.id}`}>
                  {suite.name || suite.id}
                </Link>
              </td>
              <td>{suite.tags?.join(', ')}</td>
              <td>{suite.planned_cases || ''}</td>
              <td className={suite.status === SuiteStatus.DISCONNECTED ? styles.Warn : ''}>{suite.status}</td>
              <td className={suite.result === SuiteResult.PASSED ? styles.Good : styles.Bad}>{suite.result}</td>
              <td>{new Date(suite.started_at).toISOString()}</td>
              <td>
                {!suite.finished_at
                  ? ''
                  : new Date(suite.finished_at).toISOString()}
              </td>
              <td>
                {!suite.disconnected_at
                  ? ''
                  : new Date(suite.disconnected_at).toISOString()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
