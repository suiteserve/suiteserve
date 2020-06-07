import App from './App';
import Vue from 'vue';
import VueRouter from 'vue-router';
import Vuex from 'vuex';
import Cases from './components/Cases';

Vue.use(VueRouter);
Vue.use(Vuex);

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
      count: 0,
    },
    mutations: {
      increment(state) {
        state.count++;
      },
    },
  }),
});
