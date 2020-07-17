<template>
  <TabNav title="Suites" :link-gen="genLink" :items="suites"
          :have-more="haveMore" @load-more="loadMore" :stats="{
    'Running': running,
    'Finished': finished,
  }">
    <template #item="{ item }">
      <div>
        <p>{{ item.name }}</p>
        <p class="muted">{{ formatUnix(item.started_at) }}</p>
      </div>
    </template>
  </TabNav>
</template>

<script lang="ts">
  import Vue from 'vue';
  import * as api from '@/api';
  import TabNav from '@/components/TabNav.vue';

  export default Vue.extend({
    name: 'Suites',
    computed: {
      suites(): ReadonlyArray<api.Suite> {
        return this.$store.getters.suites;
      },
      running(): number {
        return this.$store.getters.suiteAggs.running;
      },
      finished(): number {
        return this.$store.getters.suiteAggs.finished;
      },
      haveMore() {
        return this.$store.getters.moreSuites;
      },
    },
    methods: {
      loadMore() {
        this.$store.dispatch('fetchSuites');
      },
      genLink: (suiteId: string) => ({
        name: 'suite',
        params: {
          suiteId,
        },
      }),
      formatUnix: (millis: number): string =>
          new Date(millis).toLocaleString([...navigator.languages], {
            weekday: 'short',
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
          }),
    },
    components: {
      TabNav,
    },
  });
</script>
