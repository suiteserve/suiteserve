import React from 'react';
import * as api from '../../api';
import { SuiteResult, SuiteStatus } from '../../api';
import { StatusSpinner, StatusSpinnerState } from './StatusSpinner';
import styles from './SuiteNavTab.module.css';
import { NavLink } from 'react-router-dom';

export const SuiteNavTab: React.FC<{
  suite: api.Suite;
}> = ({ suite }) => {
  const startedAt = new Intl.DateTimeFormat(undefined, {
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(suite.started_at));

  let statusSpinnerState = StatusSpinnerState.NONE;
  if (suite.result === SuiteResult.PASSED) {
    statusSpinnerState = StatusSpinnerState.OK;
  } else if (suite.result === SuiteResult.FAILED) {
    statusSpinnerState = StatusSpinnerState.ERROR;
  } else if (suite.status === SuiteStatus.STARTED) {
    statusSpinnerState = StatusSpinnerState.OK;
  } else if (suite.status === SuiteStatus.DISCONNECTED) {
    statusSpinnerState = StatusSpinnerState.WARN;
  }

  return (
    <NavLink
      className={styles.SuiteNavTab}
      activeClassName={styles.Active}
      to={`/suites/${suite.id}`}
    >
      <StatusSpinner
        state={statusSpinnerState}
        running={suite.status === SuiteStatus.STARTED}
      />
      <div>
        <p>{suite.name}</p>
        <p className={styles.StartedAt}>{startedAt}</p>
      </div>
    </NavLink>
  );
};
