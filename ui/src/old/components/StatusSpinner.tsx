import React from 'react';
import classNames from 'classnames';
import styles from './StatusSpinner.module.css';

export enum StatusSpinnerState {
  NONE,
  OK,
  WARN,
  ERROR,
}

export const StatusSpinner: React.FC<{
  state: StatusSpinnerState;
  running: boolean;
}> = ({ state, running }) => {
  const className = classNames(styles.StatusSpinner, {
    [styles.Ok]: state === StatusSpinnerState.OK,
    [styles.Warn]: state === StatusSpinnerState.WARN,
    [styles.Error]: state === StatusSpinnerState.ERROR,
    [styles.Running]: running,
  });

  return (
    <div>
      <div className={className} />
    </div>
  );
};
