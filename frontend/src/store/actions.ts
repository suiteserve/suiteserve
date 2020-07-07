import {ActionTree} from 'vuex';
import * as api from '@/api';
import state from '@/store/state';

export default <ActionTree<typeof state, typeof state>>{
  async fetchSuites({state, commit}) {
    const page = await api.fetchSuitePage({fromId: state.suites.nextId, limit: 10});
    commit('insertSuitePage', page);
  },
};
