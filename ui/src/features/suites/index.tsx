import React from 'react';
import * as api from '../../api';
import { SuiteStatus } from '../../api';
import { NavLink } from 'react-router-dom';

export const Suites: React.FC = () => {
  const suites = api.SAMPLE_SUITES.sort((a, b) => {
    return b.started_at - a.started_at;
  });
  return (
    <div>
      <h1>All Suites</h1>
      <ul>
        {suites.map((suite) => (
          <li key={suite.id}>
            <NavLink to={`/suites/${suite.id}`}>{suite.name}</NavLink>:&nbsp;
            <span
              style={{
                color:
                  suite.status === SuiteStatus.STARTED
                    ? 'green'
                    : suite.status === SuiteStatus.DISCONNECTED
                    ? 'goldenrod'
                    : undefined,
              }}
            >
              {suite.status}
            </span>
            <br />
            <small>{new Date(suite.started_at).toLocaleString()}</small>
          </li>
        ))}
      </ul>
    </div>
  );
};
