<template>
  <nav class="tab-nav">
    <div class="tab-nav-header">
      <h3 class="title">{{ title }}</h3>
      <slot name="header"></slot>
    </div>
    <a class="tab" href="#" v-for="item in items.slice().reverse()" :key="item.id"
       @click="onTabClick($event, item)">
      <div class="status">
        <div class="status-icon" :class="[item.status]"></div>
      </div>
      <slot name="tab" :item="item"></slot>
    </a>
  </nav>
</template>

<script>
  export default {
    name: 'TabNav',
    props: {
      title: String,
      items: Array,
      onTabClick: Object,
    },
  };
</script>

<style scoped>
  .tab-nav {
    --padding: 0.6em;

    width: 18em;
    height: 100vh;
    overflow-y: scroll;
  }

  .tab-nav p {
    font-size: 0.75em;

    margin: 0;
  }

  .tab-nav-header {
    padding: var(--padding);
  }

  .tab-nav-header .title {
    font-size: 1em;
    font-weight: 400;

    margin: 0;
  }

  .tab-nav-header > *:not(:last-child) {
    margin-bottom: var(--padding);
  }

  .tab {
    border-top: 1px solid var(--line-color);
    border-left: 3px solid transparent;
    color: inherit;
    text-decoration: none;

    padding: var(--padding);
    padding-left: calc(var(--padding) - 3px);
    display: flex;
    align-items: center;
  }

  .tab.active {
    border-left-color: var(--highlight-color);
  }

  .tab:hover {
    background-color: var(--hover-color);
  }

  .tab > *:not(:last-child):not(.flex-spacer) {
    margin-right: var(--padding)
  }

  .tab .status-icon {
    --border-width: 4px;

    border: var(--border-width) solid var(--line-color);
    border-radius: 50%;

    transition: border var(--transition-speed);
    animation: spin var(--spin-speed) linear infinite;

    width: 1.5em;
    height: 1.5em;
  }

  .tab .status-icon.created {
    /* TODO */
    animation-play-state: paused;
  }

  .tab .status-icon.running {
    border-top-color: var(--highlight-color);
  }

  .tab .status-icon.finished {
    border-color: var(--highlight-color);
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }
</style>
