import {Store} from 'vuex';
import * as api from '@/api';
import state from '@/store/state';

export default (store: Store<typeof state>) => {
  const ws = new WebSocket('wss://localhost:8080/v1/changes');

  ws.onopen = async () => {
    await store.dispatch('fetchSuites');
  };

  ws.onmessage = async ({data}) => {
    const msg = JSON.parse(data) as api.ChangeMessage;
    if (api.isChangeMessage(msg)) {
      const change = msg.payload;
      if (api.isSuiteInsertChange(change)) {
        store.commit('insertSuite', change.doc);
      } else if (api.isSuiteUpdateChange(change)) {
        store.commit('updateSuite', change);
      } else if (api.isSuiteAggsInsertChange(change)) {
        store.commit('insertSuiteAggs', change);
      } else if (api.isSuiteAggsUpdateChange(change)) {
        store.commit('updateSuiteAggs', change);
      }
    }
  };
};
