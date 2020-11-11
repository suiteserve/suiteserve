import {
  createAsyncThunk,
  createEntityAdapter,
  createSlice,
  PayloadAction,
} from '@reduxjs/toolkit';
import * as api from '../../api';
import { RootState } from '../../app/store';

const adapter = createEntityAdapter<api.Suite>({
  sortComparer: (a, b) => {
    const diff = b.startedAt - a.startedAt;
    if (diff === 0) {
      return b.id.localeCompare(a.id);
    }
    return diff;
  },
});

export function selectSuite(
  state: RootState,
  id: api.Id
): api.Suite | undefined {
  return adapter.getSelectors().selectById(state.suites, id);
}

export function selectSuites(state: RootState): api.Suite[] {
  return adapter.getSelectors().selectAll(state.suites);
}

export const fetchOne = createAsyncThunk(
  'suites/fetchOne',
  async (id: api.Id) => await api.fetchSuite(id)
);

export const fetchPage = createAsyncThunk(
  'suites/fetchPage',
  async () => (await api.fetchSuitePage()).suites
);

const slice = createSlice({
  name: 'suites',
  initialState: adapter.getInitialState(),
  reducers: {
    inserted: (state, { payload }: PayloadAction<api.Suite>) =>
      api.onEntityInserted(adapter, state, payload),
    updated: (
      state,
      { payload }: PayloadAction<api.UpdateWatchEvent<api.Suite>>
    ) => api.onEntityUpdated(adapter, state, payload),
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchOne.fulfilled, (state, { payload }) =>
        api.onEntityInserted(adapter, state, payload)
      )
      .addCase(fetchPage.fulfilled, (state, { payload }) =>
        payload.reduce(
          (state, s) => api.onEntityInserted(adapter, state, s),
          state
        )
      );
  },
});

export const { inserted, updated } = slice.actions;

export default slice.reducer;
