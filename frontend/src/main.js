import App from './App';
import Vue from 'vue';
import VueRouter from 'vue-router';
import Vuex from 'vuex';
import Cases from './components/Cases';

Vue.use(VueRouter);
Vue.use(Vuex);

/**
 * @typedef {Object} SuiteStats
 * @property {number} runningCount
 * @property {number} finishedCount
 */

new Vue({
  el: '#app',
  render: h => h(App),
  router: new VueRouter({
    mode: 'history',
    routes: [
      {
        path: '/suites/:suiteId',
        name: 'suite',
        props: {
          cases: true,
        },
        components: {
          cases: Cases,
        },
        children: [
          {
            path: 'cases/:caseId',
            name: 'case',
          },
        ],
      },
    ],
  }),
  store: new Vuex.Store({
    state: {
      /** @type {SuiteStats} */
      suiteStats: {},
      /** @type {Suite[]} */
      suites: [],
    },
    mutations: {
      /**
       * @param state
       * @param {SuiteStats} stats
       * @param {Suite[]} suites
       */
      setSuites(state, {stats, suites}) {
        state.suiteStats = stats;
        state.suites = suites;
      },
      /**
       * @param state
       * @param {Suite} suite
       */
      saveSuite(state, suite) {
        const old = state.suites.find(s => s.id === suite.id);
        if (old) {
          if (old.status === 'running') {
            state.suiteStats.runningCount--;
          } else {
            state.suiteStats.finishedCount--;
          }
        }

        state.suites = state.suites.filter(s => s.id !== suite.id);

        if (!suite.deleted) {
          state.suites.push(suite);

          if (suite.status === 'running') {
            state.suiteStats.runningCount++;
          } else {
            state.suiteStats.finishedCount++;
          }
        }
      },
    },
    actions: {},
  }),
});
