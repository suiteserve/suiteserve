import React, { useEffect, useState } from 'react';
import * as api from '../../api';
import { Link, useParams } from 'react-router-dom';
import styles from './Logs.module.css';

export const Logs: React.FC = () => {
  const { suiteId, caseId } = useParams<{ suiteId: string; caseId: string }>();
  const [logs, setLogs] = useState([] as api.LogLine[]);

  useEffect(() => {
    new api.ServerSource().getCaseLogs(caseId).then((logs) => {
      setLogs(logs.sort((a, b) => a.idx - b.idx));
    });
  }, [caseId]);

  return (
    <div className={styles.Logs}>
      <Link to='/'>All Suites</Link> /{' '}
      <Link to={`/suites/${suiteId}`}>{suiteId}</Link>
      <h1>Logs for Case {caseId}</h1>
      <table>
        <thead>
          <tr>
            <td>Index</td>
            <td>Error</td>
            <td>Line</td>
          </tr>
        </thead>
        <tbody>
          {logs.map((logLine) => (
            <tr key={logLine.id}>
              <td>{logLine.idx}</td>
              <td>{logLine.error}</td>
              <td>{logLine.line}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
