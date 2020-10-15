import React from 'react';
import * as api from '../../api';
import { CaseResult, CaseStatus } from '../../api';
import { StatusSpinner, StatusSpinnerState } from './StatusSpinner';
import styles from './SuiteNavTab.module.css';
import { NavLink } from 'react-router-dom';

export const CaseNavTab: React.FC<{
  c: api.Case;
}> = ({ c }) => {
  const createdAt = new Intl.DateTimeFormat(undefined, {
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(c.created_at));

  let statusSpinnerState = StatusSpinnerState.NONE;
  switch (c.result) {
    case CaseResult.PASSED:
      statusSpinnerState = StatusSpinnerState.OK;
      break;
    case CaseResult.SKIPPED:
      break;
    case CaseResult.ABORTED:
    case CaseResult.ERRORED:
    case CaseResult.FAILED:
      statusSpinnerState = StatusSpinnerState.ERROR;
      break;
    default:
      if (c.status === CaseStatus.STARTED) {
        statusSpinnerState = StatusSpinnerState.OK;
      }
  }

  return (
    <NavLink
      className={styles.SuiteNavTab}
      activeClassName={styles.Active}
      to={`/suites/${c.suite_id}/cases/${c.id}`}
    >
      <StatusSpinner
        state={statusSpinnerState}
        running={c.status === CaseStatus.STARTED}
      />
      <div>
        <p>{c.name}</p>
        <p className={styles.CreatedAt}>{createdAt}</p>
      </div>
    </NavLink>
  );
};
