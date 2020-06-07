<template>
  <TabNav title="Cases" :link-gen="genLink" :items="cases" :stats="{
    'Waiting': waitingCount,
    'Running': runningCount,
    'Finished': finishedCount,
  }">
    <template #item="{ item }">
      <p class="case-name">{{ item.name }}</p>
      <div class="flex-spacer"></div>
      <p class="muted">#{{ item.num }}</p>
    </template>
  </TabNav>
</template>

<script>
  import TabNav from './TabNav';

  export default {
    name: 'Cases',
    props: {
      suiteId: String,
      caseId: String,
    },
    data() {
      return {
        cases: [],
      };
    },
    created() {
      this.load();
    },
    watch: {
      $route: function (to, from) {
        if (to.params.suiteId !== from.params.suiteId) {
          this.load();
        }
      },
    },
    computed: {
      waitingCount() {
        return this.cases.filter(c => c.status === 'created').length;
      },
      runningCount() {
        return this.cases.filter(c => c.status === 'running').length;
      },
      finishedCount() {
        return this.cases
          .filter(c => c.status !== 'created' && c.status !== 'running')
          .length;
      },
    },
    methods: {
      genLink(caseId) {
        return {
          name: 'case',
          params: {
            suiteId: this.suiteId,
            caseId,
          },
        };
      },
      async load() {
        const suiteId = this.suiteId;
        const url = new URL(`/v1/suites/${suiteId}/cases`, window.location.href);
        const res = await fetch(url.href);
        const json = await res.json();

        if (!res.ok) {
          throw `Error loading cases: ${json.error}`;
        }

        // this.suiteId could have changed while awaiting
        if (this.suiteId === suiteId) {
          this.cases = json;
        }
      },
    },
    components: {
      TabNav,
    },
  };
</script>

<style scoped>
  .case-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
