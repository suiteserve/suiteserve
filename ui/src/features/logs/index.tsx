import React, {useEffect, useState} from 'react';
import * as api from '../../api';
import { Link, useParams } from 'react-router-dom';
import styles from './Logs.module.css';
import {useDispatch, useSelector} from 'react-redux';
import {fetchForCase, selectLogs} from './slice';

export const Logs: React.FC = () => {
  const { suiteId, caseId } = useParams<{ suiteId: api.Id; caseId: api.Id }>();
  const dispatch = useDispatch();
  const logs = useSelector(selectLogs);

  useEffect(() => {
    dispatch(fetchForCase(caseId));
  }, [dispatch, caseId]);

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
