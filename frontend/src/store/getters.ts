import {GetterTree} from 'vuex';
import * as api from '@/api';
import state from '@/store/state';

export default <GetterTree<typeof state, typeof state>>{
  suites(state): ReadonlyArray<api.Suite> {
    return state.suites.entities
        .filter(s => !s.deleted)
        .concat()
        .sort((a, b) => {
          if (a.started_at != b.started_at) {
            return a.started_at - b.started_at;
          }
          // assert a.id =/= b.id
          return a.id < b.id ? -1 : 1;
        })
        .reverse();
  },
  suiteAggs(state): api.SuiteAggs {
    return state.suites.aggs.entities[0] || {
      id: 'suite_aggs',
      version: 0,
      running: 0,
      finished: 0,
    };
  },
  moreSuites(state): boolean {
    return !state.suites.init
        || state.suites.nextId != null
        || state.suites.pendingUpdates.length > 0;
  },
};
