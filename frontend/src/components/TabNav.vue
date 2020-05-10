<template>
  <nav>
    <header>
      <h3 class="title">{{ title }}</h3>
      <slot name="header"></slot>
      <div class="stats">
        <p v-for="(value, name) in stats">
          <span class="muted">{{ name }}</span> {{ value }}
        </p>
      </div>
    </header>
    <a class="tab" href="#" v-for="item in items.slice().reverse()" :key="item.id"
       @click="openTab($event, item)" :class="{ active: item.id === activeTabId }">
      <div>
        <div class="status-icon" :class="item.status"></div>
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
      stats: Object,
      items: Array,
    },
    data() {
      return {
        activeTabId: undefined,
      };
    },
    methods: {
      openTab(event, item) {
        event.preventDefault();
        this.activeTabId = item.id;
        this.$emit('open-tab', item)
      }
    },
  };
</script>

<style scoped>
  nav {
    --padding: 0.6em;

    width: 18em;
    height: 100vh;
    overflow-y: scroll;
  }

  p {
    font-size: 0.75em;

    margin: 0;
  }

  header {
    padding: var(--padding);
  }

  .title {
    font-size: 1em;
    font-weight: 400;

    margin: 0;
  }

  header > *:not(:last-child) {
    margin-bottom: var(--padding);
  }

  .stats {
    display: flex;
  }

  .stats > *:not(:last-child) {
    margin-right: 1em;
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

  .status-icon {
    --border-width: 4px;

    border: var(--border-width) solid var(--line-color);
    border-radius: 50%;

    transition: border var(--transition-speed);
    animation: spin var(--spin-speed) linear;

    width: 1.5em;
    height: 1.5em;
  }

  .status-icon.running {
    border-top-color: var(--highlight-color);

    animation-iteration-count: infinite;
  }

  .status-icon.finished {
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
