import * as api from '@/api';

export default {
  suites: {
    init: false,
    pendingUpdates: [] as ReadonlyArray<api.UpdateChange<api.Suite>>,
    entities: [] as ReadonlyArray<api.Suite>,
    nextId: undefined as string | undefined,
    aggs: {
      pendingUpdates: [] as ReadonlyArray<api.UpdateChange<api.SuiteAggs>>,
      entities: [] as ReadonlyArray<api.SuiteAggs>,
    },
  },
};
