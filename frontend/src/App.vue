<template>
  <main id="app">
    <TabNav :items="suites" :onTabClick="openSuite" title="Suites">
      <template #header>
        <div class="suites-stats">
          <p>
            <span class="muted">Running:</span> <span v-text="runningSuites"></span>
          </p>
          <p>
            <span class="muted">Total:</span> <span v-text="suites.length"></span>
          </p>
        </div>
      </template>
      <template #tab="{ item }">
        <div>
          <p v-text="formatTime(item.created_at)"></p>
          <p class="muted" v-text="item.id"></p>
        </div>
      </template>
    </TabNav>

    <TabNav :items="cases" title="Cases">
      <template #tab="{ item }">
        <div class="case-name">
          <p v-text="item.name"></p>
        </div>
        <div class="flex-spacer"></div>
        <div class="case-num">
          <p class="muted">#<span v-text="item.num"></span></p>
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

  let activeSuiteElem;

  export default {
    name: 'App',
    created() {
      retry.bind(this)(() => true, fetchSuites)
        .then(suites => this.suites = suites);
    },
    data() {
      return {
        cases: [],
        runningSuites: 0,
        suites: [],
      };
    },
    methods: {
      formatTime,
      openSuite: function (event, suite) {
        event.preventDefault();
        const e = event.currentTarget;

        if (activeSuiteElem) {
          activeSuiteElem.classList.remove('active');
        }
        activeSuiteElem = e;
        activeSuiteElem.classList.add('active');

        retry.bind(this)(() => true, fetchCases, suite.id)
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
    color: #6c6c80;
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

  .tab-nav-header .suites-stats {
    display: flex;
  }

  .tab-nav-header .suites-stats > *:not(:last-child):not(.flex-spacer) {
    margin-right: 1em;
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
