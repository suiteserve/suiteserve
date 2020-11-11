import { combineReducers, configureStore } from '@reduxjs/toolkit';
import cases from '../features/cases/slice';
import logs from '../features/logs/slice';
import suites from '../features/suites/slice';
import { newWatchEventsMiddleware } from '../api';

const rootReducer = combineReducers({
  cases,
  logs,
  suites,
});

export type RootState = ReturnType<typeof rootReducer>;

export const store = configureStore({
  reducer: rootReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().prepend(newWatchEventsMiddleware()),
});

export type AppDispatch = typeof store.dispatch;
