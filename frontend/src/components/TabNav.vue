<template>
  <nav>
    <header>
      <h3 class="title">{{ title }}</h3>
      <div class="stats">
        <p v-for="(value, key) in stats">
          <span class="muted">{{ key }}</span> {{ value }}
        </p>
      </div>
      <slot name="header"></slot>
    </header>
    <div class="items">
      <a class="tab" href="#" v-for="item in items" :key="item.id"
         @click="openTab($event, item)">
        <div class="inner-tab" :class="{ active: item.id === activeTabId }">
          <div class="status-icon-container">
            <div class="status-icon" :class="`status-${item.status}`"></div>
          </div>
          <slot name="tab" :item="item"></slot>
        </div>
      </a>
      <a class="tab" href="#" v-if="more" @click="loadMore">
        <div class="inner-tab">
          <p>Load More</p>
        </div>
      </a>
    </div>
  </nav>
</template>

<script>
  export default {
    name: 'TabNav',
    props: {
      title: String,
      stats: Object,
      items: Array,
      more: Boolean,
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
        this.$emit('open-tab', item);
      },
      loadMore(event) {
        event.preventDefault();
        this.$emit('load-more');
      },
    },
  };
</script>

<style scoped>
  nav {
    --status-passed-color: #52af5c;
    --status-failed-color: #af525a;
    --padding: 10px;

    width: max-content;
    max-width: 25em;
    height: 100vh;

    display: flex;
    flex-direction: column;
  }

  p {
    margin: 0;
  }

  header {
    padding: var(--padding);
  }

  header > *:not(:last-child) {
    margin-bottom: var(--padding);
  }

  .title {
    font-size: 1rem;
    font-weight: 400;

    margin: 0;
  }

  .stats {
    display: flex;
  }

  .stats > *:not(:last-child) {
    margin-right: 1em;
  }

  .items {
    overflow-y: scroll;
  }

  .tab {
    border-top: 1px solid var(--line-color);
    color: inherit;
    text-decoration: none;

    display: block;
  }

  .tab:hover {
    background-color: var(--hover-color);
  }

  .inner-tab {
    border-left: 3px solid transparent;

    padding: var(--padding);
    padding-left: calc(var(--padding) - 3px);
    display: flex;
    align-items: center;
  }

  .inner-tab.active {
    border-left-color: var(--highlight-color);
  }

  .inner-tab > *:not(.flex-spacer):not(:last-child) {
    margin-right: var(--padding)
  }

  .status-icon {
    --border-width: 0.25em;
    --size: 1em;

    border-radius: 50%;

    transition: border var(--transition-speed);

    width: var(--size);
    height: var(--size);
    margin: var(--border-width);
    box-sizing: content-box;
  }

  .status-icon.status-created {
    border: var(--border-width) solid var(--line-color);

    margin: 0;
  }

  .status-icon.status-disabled, .status-icon.status-disconnected {
    background-color: var(--line-color);
  }

  .status-icon.status-running {
    border: var(--border-width) solid var(--line-color);
    border-top-color: var(--highlight-color);

    animation: spin var(--spin-speed) linear infinite;

    margin: 0;
  }

  .status-icon.status-passed {
    background-color: var(--status-passed-color);
  }

  .status-icon.status-failed, .status-icon.status-errored {
    background-color: var(--status-failed-color);
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
