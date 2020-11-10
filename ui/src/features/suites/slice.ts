import {
  createAsyncThunk,
  createEntityAdapter,
  createSlice,
  PayloadAction,
} from '@reduxjs/toolkit';
import * as api from '../../api';
import { State } from '../../app/store';

const adapter = createEntityAdapter<api.Suite>({
  sortComparer: (a, b) => {
    const timeDiff = b.startedAt - a.startedAt;
    if (timeDiff === 0) {
      return b.id.localeCompare(a.id);
    }
    return timeDiff;
  },
});

export function selectSuite(state: State, id: api.Id): api.Suite | undefined {
  return adapter.getSelectors().selectById(state.suites, id);
}

export function selectSuites(state: State): api.Suite[] {
  return adapter.getSelectors().selectAll(state.suites);
}

export const fetchOne = createAsyncThunk(
  'suites/fetchOne',
  async (id: api.Id) => await api.getSuite(id)
);

export const fetchPage = createAsyncThunk(
  'suites/fetchPage',
  async () => (await api.getSuitePage()).suites
);

const slice = createSlice({
  name: 'suites',
  initialState: adapter.getInitialState(),
  reducers: {
    update(state, { payload }: PayloadAction<api.WatchEvent<api.Suite>>) {
      api.applyWatchEvent(adapter, state, payload);
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchOne.fulfilled, (state, { payload }) => {
        api.upsertEntity(adapter, state, payload);
      })
      .addCase(fetchPage.fulfilled, (state, { payload }) => {
        api.upsertEntities(adapter, state, payload);
      });
  },
});

const { update } = slice.actions;

export default slice.reducer;

export const applyWatchEvent = createAsyncThunk<
  void,
  api.WatchEvent<api.Suite>,
  {
    state: State;
  }
>('suites/applyWatchEvent', (evt, { getState, dispatch }) => {
  if (selectSuite(getState(), evt.id)) {
    dispatch(update(evt));
  } else {
    dispatch(fetchOne(evt.id));
  }
});
