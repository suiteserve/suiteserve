<template>
  <main id="app">
    <TabNav title="Suites" :items="suites" :stats="{
      'Running': runningSuites,
      'Finished': finishedSuites,
    }" @open-tab="openSuite">
      <template #tab="{ item }">
        <div>
          <p>{{ formatTime(item.created_at) }}</p>
          <p class="muted">{{ item.id }}</p>
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
  import {formatTime, retry} from './util';
  import TabNav from './components/TabNav';
  import {fetchCases} from './cases';

  export default {
    name: 'App',
    created() {
      retry.bind(this)(() => true, fetchSuites)
        .then(suites => this.suites = suites);
    },
    data() {
      return {
        cases: [],
        suites: [],
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
      runningSuites() {
        return this.suites.filter(s => s.status === 'running').length;
      },
      finishedSuites() {
        return this.suites
          .filter(s => s.status !== 'created' && s.status !== 'running')
          .length;
      },
    },
    methods: {
      formatTime,
      openCase(c) {
        console.log('TODO');
      },
      openSuite(s) {
        retry.bind(this)(() => true, fetchCases, s.id)
          .then(cases => this.cases = cases)
          .catch(() => {
          });
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
