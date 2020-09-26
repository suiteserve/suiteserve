import { createRouter, createWebHistory } from 'vue-router'

export default createRouter({
  history: createWebHistory(),
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
});
