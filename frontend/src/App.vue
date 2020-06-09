<template>
  <main id="app">
    <Suites/>
    <router-view name="cases"></router-view>
  </main>
</template>

<script>
  import Suites from './components/Suites';
  import * as api from './api';
  import {mapMutations} from 'vuex';

  export default {
    name: 'App',
    data() {
      return {

      };
    },
    created() {
      const events = new EventSource('/v1/events');

      events.onmessage = e => {
        const json = JSON.parse(e.data);

        if (json.coll === 'suites') {
          this.saveSuite(json.payload);
        }
      };

      events.onopen = async () => {
        const page = await api.fetchSuites({});

        this.setSuites({
          stats: {
            runningCount: page.running_count,
            finishedCount: page.finished_count,
          },
          suites: page.suites,
        });
      };
    },
    methods: {
      ...mapMutations([
        'saveSuite',
        'setSuites',
      ]),
    },
    components: {
      Suites,
    },
  };
</script>

<style>
  #app {
    display: flex;
  }
</style>
