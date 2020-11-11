import {
  createAsyncThunk,
  createEntityAdapter,
  createSelector,
  createSlice,
  PayloadAction,
} from '@reduxjs/toolkit';
import * as api from '../../api';
import { RootState } from '../../app/store';

const adapter = createEntityAdapter<api.Case>({
  sortComparer: (a, b) => a.idx - b.idx,
});

export function selectCase(state: RootState, id: api.Id): api.Case | undefined {
  return adapter.getSelectors().selectById(state.cases.entities, id);
}

export const selectCases = createSelector(
  (state: RootState) => state.cases,
  (cases) =>
    adapter
      .getSelectors()
      .selectAll(cases.entities)
      .filter((e) => e.suiteId === cases.suiteId)
);

export const fetchOne = createAsyncThunk(
  'cases/fetchOne',
  async (id: api.Id) => await api.fetchCase(id)
);

export const fetchForSuite = createAsyncThunk(
  'cases/fetchForSuite',
  async (suiteId: api.Id) => await api.fetchSuiteCases(suiteId)
);

const slice = createSlice({
  name: 'cases',
  initialState: {
    suiteId: '' as api.Id,
    entities: adapter.getInitialState(),
  },
  reducers: {
    inserted(state, { payload }: PayloadAction<api.Case>) {
      api.onEntityInserted(adapter, state.entities, payload);
    },
    updated(state, { payload }: PayloadAction<api.UpdateWatchEvent<api.Case>>) {
      api.onEntityUpdated(adapter, state.entities, payload);
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchOne.fulfilled, (state, { payload }) => {
        api.onEntityInserted(adapter, state.entities, payload);
      })
      .addCase(fetchForSuite.fulfilled, (state, { payload, meta }) => {
        state.suiteId = meta.arg;
        payload.forEach((s) => {
          api.onEntityInserted(adapter, state.entities, s);
        });
      });
  },
});

export const { inserted, updated } = slice.actions;

export default slice.reducer;
