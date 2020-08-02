<template>
  <main id="app">
    <Suites/>
    <router-view name="cases"></router-view>
  </main>
</template>

<script>
import Suites from '@/components/Suites';

export default {
  name: 'App',
  components: {
    Suites,
  },
}
</script>

<style>
#app {
  display: flex;
}
</style>
