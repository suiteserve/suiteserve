import React, {useState, useEffect} from 'react';
import * as api from '../../api';
import { Link, useParams } from 'react-router-dom';
import styles from './Logs.module.css';

export const Logs: React.FC = () => {
  const { suiteId, caseId } = useParams<{ suiteId: string; caseId: string }>();
  const [logs, setLogs] = useState([] as api.LogLine[])

  useEffect(() => {
    new api.ServerSource().getCaseLogs(caseId).then(logs => {
      setLogs(logs.sort((a, b) => {
        if (a.idx === b.idx) {
          return b.idx - a.idx;
        }
        return b.timestamp - a.timestamp;
      }))
    })
  }, [caseId])

  return (
    <div className={styles.Logs}>
      <Link to='/'>All Suites</Link> / <Link to={`/suites/${suiteId}`}>{suiteId}</Link>
      <h1>Logs for Case {caseId}</h1>
      <table>
        <thead>
          <tr>
            <td>Index</td>
            <td>Level</td>
            <td>Trace</td>
            <td>Message</td>
            <td>Timestamp</td>
          </tr>
        </thead>
        <tbody>
          {logs.map((logLine) => (
            <tr key={logLine.id}>
              <td>{logLine.idx}</td>
              <td>{logLine.level}</td>
              <td><pre>{logLine.trace}</pre></td>
              <td><pre>{logLine.message}</pre></td>
              <td>{new Date(logLine.timestamp).toISOString()}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
