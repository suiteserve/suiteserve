import React from 'react';
import { Suites } from './features/suites';
import { Cases } from './features/cases';
import { Logs } from './features/logs';
import { Route, Switch } from 'react-router-dom';

export const App: React.FC = () => {
  return (
    <div>
      <Switch>
        <Route path='/suites/:suiteId/cases/:caseId' component={Logs} />
        <Route path='/suites/:suiteId' component={Cases} />
        <Route path='/' component={Suites} />
      </Switch>
    </div>
  );
};
