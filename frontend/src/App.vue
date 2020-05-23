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
        <div class="case-name">
          <p>{{ item.name }}</p>
        </div>
        <div class="flex-spacer"></div>
        <div class="case-num">
          <p class="muted">#{{ item.num }}</p>
        </div>
      </template>
    </TabNav>
  </main>
</template>

<script>
  import {fetchSuites} from './suites';
  import {formatTime} from './util';
  import TabNav from './components/TabNav';
  import {fetchCases} from './cases';

  export default {
    name: 'App',
    async created() {
      const suitesRes = await fetchSuites(null, 10)
      this.suites = suitesRes.suites
      this.runningSuites = suitesRes.running
      this.finishedSuites = suitesRes.finished
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
      formatTime,
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
        this.runningSuites = suitesRes.running
        this.finishedSuites = suitesRes.finished
        this.nextId = suitesRes.next_id
      },
    },
    components: {
      TabNav,
    },
  };
</script>

<style>
  * {
    box-sizing: border-box;
  }

  :root {
    --bg-color: #23232d;
    --hover-color: #282833;
    --line-color: #3c3c4d;
    --muted-color: #6c6c80;
    --highlight-color: #9f9fcc;

    --spin-speed: 1.5s;
    --transition-speed: 0.3s;

    scrollbar-color: var(--line-color) transparent;
    scrollbar-width: thin;
  }

  ::-webkit-scrollbar {
    width: 12px;
  }

  ::-webkit-scrollbar-track {
    background: var(--bg-color);
  }

  ::-webkit-scrollbar-thumb {
    background-color: var(--line-color);
    border: 3px solid var(--bg-color);
    border-radius: 6px;
  }

  .muted {
    color: var(--muted-color);
  }

  .flex-spacer {
    flex-grow: 1;
  }

  body {
    background-color: var(--bg-color);
    color: #fff;
    font: 400 1em/1.3 'Fira Mono', monospace;

    margin: 0;
  }

  noscript {
    height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  #app {
    display: flex;
  }

  .case-name {
    overflow: hidden;
  }

  .case-name p {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .case-num {
    width: max-content;
  }
</style>
