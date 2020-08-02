import Vue from 'vue';
import VueRouter from 'vue-router';
import Cases from '@/components/Cases';

Vue.use(VueRouter);

export default new VueRouter({
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
})
