import React from 'react';
import * as api from '../../api';
import { SuiteNavTab } from './SuiteNavTab';
import styles from './SuiteNav.module.css';

export const SuiteNav: React.FC<{
  suites: api.Suite[];
}> = ({ suites }) => {
  return (
    <nav className={styles.SuiteNav}>
      {suites.map((suite) => (
        <SuiteNavTab key={suite.id} suite={suite} />
      ))}
    </nav>
  );
};
