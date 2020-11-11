import {
  createAsyncThunk,
  createEntityAdapter,
  createSelector,
  createSlice,
  PayloadAction,
} from '@reduxjs/toolkit';
import * as api from '../../api';
import { RootState } from '../../app/store';

const adapter = createEntityAdapter<api.LogLine>({
  sortComparer: (a, b) => a.idx - b.idx,
});

export function selectLogLine(state: RootState, id: api.Id): api.LogLine | undefined {
  return adapter.getSelectors().selectById(state.logs.entities, id);
}

export const selectLogs = createSelector(
  (state: RootState) => state.logs,
  (logs) =>
    adapter
      .getSelectors()
      .selectAll(logs.entities)
      .filter((e) => e.caseId === logs.caseId)
);

export const fetchOne = createAsyncThunk(
  'logs/fetchOne',
  async (id: api.Id) => await api.fetchLogLine(id)
);

export const fetchForCase = createAsyncThunk(
  'logs/fetchForCase',
  async (caseId: api.Id) => await api.fetchCaseLogs(caseId)
);

const slice = createSlice({
  name: 'logs',
  initialState: {
    caseId: '' as api.Id,
    entities: adapter.getInitialState(),
  },
  reducers: {
    inserted(state, { payload }: PayloadAction<api.LogLine>) {
      api.onEntityInserted(adapter, state.entities, payload);
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchOne.fulfilled, (state, { payload }) => {
        api.onEntityInserted(adapter, state.entities, payload);
      })
      .addCase(fetchForCase.fulfilled, (state, { payload, meta }) => {
        state.caseId = meta.arg;
        payload.forEach((s) => {
          api.onEntityInserted(adapter, state.entities, s);
        });
      });
  },
});

export const { inserted } = slice.actions;

export default slice.reducer;
