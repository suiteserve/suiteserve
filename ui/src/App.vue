<template>
  <HelloWorld msg="Hello Vue 3.0 + Vite" />
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import HelloWorld from './components/HelloWorld.vue';
import * as api from './api';

export default defineComponent({
  name: 'App',
  components: {
    HelloWorld,
  },
  async mounted(): Promise<void> {
    console.log((await api.getAllAttachments()).map((value) => value.timestamp));
  },
});
</script>
