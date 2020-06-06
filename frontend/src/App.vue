<template>
  <main id="app">
    <TabNav title="Suites" :items="suites" :more="nextId != null" :stats="{
      'Running': runningSuites,
      'Finished': finishedSuites,
    }" @open-tab="openSuite" @load-more="loadMoreSuites">
      <template #tab="{ item }">
        <div>
          <p>{{ item.name }}</p>
          <p class="muted">{{ formatTime(item.started_at) }}</p>
        </div>
      </template>
    </TabNav>

    <TabNav title="Cases" :items="cases" :stats="{
      'Waiting': waitingCases,
      'Running': runningCases,
      'Finished': finishedCases,
    }" @open-tab="openCase">
      <template #tab="{ item }">
        <p class="case-name">{{ item.name }}</p>
        <div class="flex-spacer"></div>
        <p class="muted">#{{ item.num }}</p>
      </template>
    </TabNav>
  </main>
</template>

<script>
  import {fetchSuites} from './suites';
  import {formatUnix} from './util';
  import TabNav from './components/TabNav';
  import {fetchCases} from './cases';

  export default {
    name: 'App',
    async created() {
      const suitesRes = await fetchSuites(null, 10)
      this.suites = suitesRes.suites
      this.runningSuites = suitesRes.running_count
      this.finishedSuites = suitesRes.finished_count
      this.nextId = suitesRes.next_id
    },
    data() {
      return {
        cases: [],
        suites: [],
        runningSuites: 0,
        finishedSuites: 0,
        nextId: null,
      };
    },
    computed: {
      waitingCases() {
        return this.cases.filter(c => c.status === 'created').length;
      },
      runningCases() {
        return this.cases.filter(c => c.status === 'running').length;
      },
      finishedCases() {
        return this.cases
          .filter(c => c.status !== 'created' && c.status !== 'running')
          .length;
      },
    },
    methods: {
      formatTime: formatUnix,
      openCase(c) {
        console.log('TODO');
      },
      async openSuite(s) {
        this.cases = await fetchCases(s.id)
      },
      async loadMoreSuites() {
        let suitesRes;
        if (this.suites.length) {
          suitesRes = await fetchSuites(this.nextId, 10);
        } else {
          suitesRes = await fetchSuites(null, 10);
        }

        this.suites.push(...suitesRes.suites)
        this.runningSuites = suitesRes.running_count
        this.finishedSuites = suitesRes.finished_count
        this.nextId = suitesRes.next_id
      },
    },
    components: {
      TabNav,
    },
  };
</script>

<style>
  #app {
    display: flex;
  }

  .case-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
