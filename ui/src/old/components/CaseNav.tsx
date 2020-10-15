import React from 'react';
import * as api from '../../api';
import { CaseNavTab } from './CaseNavTab';
import styles from './CaseNav.module.css';

export const CaseNav: React.FC<{
  cases: api.Case[];
}> = ({ cases }) => {
  return (
    <nav className={styles.CaseNav}>
      {cases.map((c) => (
        <CaseNavTab key={c.id} c={c} />
      ))}
    </nav>
  );
};
