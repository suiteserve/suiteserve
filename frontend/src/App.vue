<template>
  <main id="app">
    <Suites/>
    <router-view name="cases"></router-view>
  </main>
</template>

<script lang="ts">
  import Suites from '@/components/Suites.vue';

  export default {
    name: 'App',
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
