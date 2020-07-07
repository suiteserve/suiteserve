<template>
  <TabNav title="Cases" :link-gen="genLink" :items="[]" :stats="{
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

<script lang="ts">
  import Vue from 'vue';
  import {Route} from 'vue-router';
  import TabNav from '@/components/TabNav.vue';

  export default Vue.extend({
    name: 'Cases',
    props: {
      suiteId: String,
      caseId: String,
    },
    watch: {
      $route: function (to: Route, from: Route) {
        if (to.params.suiteId !== from.params.suiteId) {
          // this.load();
        }
      },
    },
    computed: {
      waitingCount(): number {
        return 3;
      },
      runningCount(): number {
        return 5;
      },
      finishedCount(): number {
        return 9;
      },
    },
    methods: {
      genLink(caseId: string) {
        return {
          name: 'case',
          params: {
            suiteId: this.suiteId,
            caseId,
          },
        };
      },
    },
    components: {
      TabNav,
    },
  });
</script>

<style scoped>
  .case-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
