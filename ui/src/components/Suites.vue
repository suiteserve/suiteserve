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
import rpc from '@/api';

export default {
  name: 'Suites',
  data() {
    return {
      suites: [],
      running: 0,
      total: 0,
    };
  },
  computed: {
    finished() {
      return this.total - this.running;
    },
  },
  mounted() {
    const req = new rpc.WatchSuitesRequest();
    req.setId('');
    req.setPadOlder(10);

    const stream = rpc.watchSuites(req);
    stream.on('data', resp => {
      console.log(resp);
    });
    stream.on('end', () => {
      console.log('end');
    });
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
};
</script>
