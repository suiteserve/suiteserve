import React, {useContext, useEffect, useState} from 'react';
import * as api from '../../api';
import { SuiteResult, SuiteStatus } from '../../api';
import styles from './Suites.module.css';
import { Link } from 'react-router-dom';

export const Suites: React.FC = () => {
  const apiSource = useContext(api.APIContext);
  const [suites, setSuites] = useState([] as api.Suite[]);

  useEffect(() => {
    apiSource.getSuitePage().then((page) => {
      setSuites(page.suites.sort((a, b) => {
        return b.startedAt - a.startedAt;
      }));
    });
    apiSource.watch('suites', (evt: api.WatchEvent<api.Suite>) => {
      setSuites(suites => {
        const newSuites = api.applyWatchEvent(evt)(suites);
        if (newSuites === undefined) {
          apiSource.getSuite(evt.id).then(s => {
            setSuites(suites => suites.concat(s));
          });
          return suites;
        }
        return newSuites;
      })
    });
    return () => apiSource.unwatch('suites');
  }, [apiSource]);

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
                <Link to={`/suites/${suite.id}`}>{suite.name || suite.id}</Link>
              </td>
              <td>{suite.tags?.join(', ')}</td>
              <td>{suite.plannedCases || ''}</td>
              <td
                className={
                  suite.status === SuiteStatus.DISCONNECTED ? styles.Warn : ''
                }
              >
                {suite.status}
              </td>
              <td
                className={
                  suite.result === SuiteResult.PASSED ? styles.Good : styles.Bad
                }
              >
                {suite.result}
              </td>
              <td>{new Date(suite.startedAt).toISOString()}</td>
              <td>
                {!suite.finishedAt
                  ? ''
                  : new Date(suite.finishedAt).toISOString()}
              </td>
              <td>
                {!suite.disconnectedAt
                  ? ''
                  : new Date(suite.disconnectedAt).toISOString()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
