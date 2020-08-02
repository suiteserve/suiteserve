<template>
  <TabNav :items="suites" :link-gen="genLink" :stats="{
    'Running': running,
    'Finished': finished,
  }" title="Suites">
    <template #item="{ item }">
      <div>
        <p>{{ item.name }}</p>
        <p class="muted">{{ formatUnix(item.started_at) }}</p>
      </div>
    </template>
  </TabNav>
</template>

<script>
import TabNav from '@/components/TabNav';

export default {
  name: 'Suites',
  computed: {
    suites() {
      return [
        {
          id: "132",
          status: 'running',
          name: "Hello Hi there",
          started_at: 0,
        }
      ];
    },
    running() {
      return 5;
    },
    finished() {
      return 9;
    },
  },
  methods: {
    loadMore() {

    },
    /**
     * @param {string} suiteId
     * @return {Object}
     */
    genLink(suiteId) {
      return {
        name: 'suite',
        params: {
          suiteId,
        },
      };
    },
    /**
     * @param {number} millis
     * @return {string}
     */
    formatUnix(millis) {
      return new Date(millis).toLocaleString([...navigator.languages], {
        weekday: 'short',
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });
    },
  },
  components: {
    TabNav,
  },
}
</script>
