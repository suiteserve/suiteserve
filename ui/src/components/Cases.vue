<template>
  <TabNav :items="[]" :link-gen="genLink" :stats="{
    'Waiting': waitingCount,
    'Running': runningCount,
    'Finished': finishedCount,
  }" title="Cases">
    <template #item="{ item }">
      <p class="case-name">{{ item.name }}</p>
      <div class="flex-spacer"></div>
      <p class="muted">#{{ item.num }}</p>
    </template>
  </TabNav>
</template>

<script>
import TabNav from '@/components/TabNav';

export default {
  name: 'Cases',
  props: {
    suiteId: String,
    caseId: String,
  },
  watch: {
    $route(to, from) {
      if (to.params.suiteId !== from.params.suiteId) {
        // this.load();
      }
    },
  },
  computed: {
    waitingCount() {
      return 3;
    },
    runningCount() {
      return 5;
    },
    finishedCount() {
      return 9;
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
  },
  components: {
    TabNav,
  },
}
</script>

<style scoped>
.case-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
