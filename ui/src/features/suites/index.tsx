import React, {useEffect} from 'react';
import {SuiteResult, SuiteStatus} from '../../api';
import styles from './Suites.module.css';
import {Link} from 'react-router-dom';
import {useDispatch, useSelector} from 'react-redux';
import {fetchPage, selectSuites} from './slice';

export const Suites: React.FC = () => {
  const dispatch = useDispatch();
  const suites = useSelector(selectSuites);

  useEffect(() => {
    dispatch(fetchPage());
  }, [dispatch]);

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
