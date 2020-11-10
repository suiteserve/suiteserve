import React, {useEffect} from 'react';
import {Suites} from './features/suites';
import {Cases} from './features/cases';
import {Logs} from './features/logs';
import {Route, Switch} from 'react-router-dom';
import {useDispatch} from 'react-redux';
import * as api from './api';

export const App: React.FC = () => {
  const dispatch = useDispatch();

  useEffect(() => {
    dispatch(api.watchChanges());
  }, [dispatch]);

  return (
    <Switch>
      <Route path='/suites/:suiteId/cases/:caseId' component={Logs}/>
      <Route path='/suites/:suiteId' component={Cases}/>
      <Route path='/' component={Suites}/>
    </Switch>
  );
};
