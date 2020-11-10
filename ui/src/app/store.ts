import {configureStore} from '@reduxjs/toolkit';
import suites from '../features/suites/slice';

export const store = configureStore({
  reducer: {
    suites,
  },
});

export type State = ReturnType<typeof store.getState>;
