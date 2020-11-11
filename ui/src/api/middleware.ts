import { Middleware, MiddlewareAPI } from '@reduxjs/toolkit';
import * as cases from '../features/cases/slice';
import * as logs from '../features/logs/slice';
import * as suites from '../features/suites/slice';
import * as t from './types';
import { AppDispatch, RootState } from '../app/store';

export const newWatchEventsMiddleware = (): Middleware => (
  store: MiddlewareAPI<AppDispatch, RootState>
) => {
  const sse = new EventSource('/v1/suites?watch=true');
  const on = <E extends t.Watchable>(
    coll: string,
    handler: (evt: t.WatchEvent<E>) => void
  ) => {
    sse.addEventListener(coll, ((evt: MessageEvent) => {
      handler(JSON.parse(evt.data));
    }) as EventListener);
  };
  on('cases', (evt: t.WatchEvent<t.Case>) => {
    if (t.isInsertWatchEvent(evt)) {
      store.dispatch(cases.inserted(evt.insert));
    } else if (
      t.isUpdateWatchEvent(evt) &&
      cases.selectCase(store.getState(), evt.id)
    ) {
      store.dispatch(cases.updated(evt));
    } else {
      store.dispatch(cases.fetchOne(evt.id));
    }
  });
  on('logs', (evt: t.WatchEvent<t.LogLine>) => {
    if (t.isInsertWatchEvent(evt)) {
      store.dispatch(logs.inserted(evt.insert));
    }
  });
  on('suites', (evt: t.WatchEvent<t.Suite>) => {
    if (t.isInsertWatchEvent(evt)) {
      store.dispatch(suites.inserted(evt.insert));
    } else if (
      t.isUpdateWatchEvent(evt) &&
      suites.selectSuite(store.getState(), evt.id)
    ) {
      store.dispatch(suites.updated(evt));
    } else {
      store.dispatch(suites.fetchOne(evt.id));
    }
  });
  return (next) => (action) => next(action);
};
