<template>
  <h1>{{ msg }}</h1>
  <img
    alt="Vue logo"
    src="../assets/logo.png"
  >
  <button @click="count++">
    Count is: {{ count }}
  </button>
  <p>Edit <code>components/HelloWorld.vue</code> to test hot module replacement.</p>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'HelloWorld',
  props: {
    msg: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      count: 0,
    };
  },
});
</script>
