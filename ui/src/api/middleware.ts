import {createAsyncThunk} from '@reduxjs/toolkit';
import * as t from './types';
import {applyWatchEvent as applySuiteWatchEvent} from '../features/suites/slice';
import {State} from '../app/store';

export const watchChanges = createAsyncThunk<void,
  void,
  {
    state: State;
  }>('watchChanges', async (_, {dispatch}) => {
  const sse = new EventSource('/v1/suites?watch=true');

  const on = <E extends t.Watchable>(
    coll: string,
    handler: (evt: t.WatchEvent<E>) => void,
  ) =>
    sse.addEventListener(coll, ((evt: MessageEvent) =>
      handler(JSON.parse(evt.data))) as EventListener);

  on('suites', (evt: t.WatchEvent<t.Suite>) => {
    dispatch(applySuiteWatchEvent(evt));
  });
});
