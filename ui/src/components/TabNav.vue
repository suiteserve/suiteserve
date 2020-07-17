<template>
  <nav class="nav">
    <header class="nav-header">
      <h3 class="nav-header-title">{{ title }}</h3>
      <div class="nav-header-stats">
        <p v-for="(value, key) in stats">
          <span class="muted">{{ key }}</span> {{ value }}
        </p>
      </div>
      <slot name="header"></slot>
    </header>
    <div class="nav-items">
      <router-link :to="linkGen(item.id)" class="nav-item"
                   v-for="item in items" :key="item.id">
        <div class="nav-item-inner">
          <div class="status-icon-container">
            <div class="status-icon" :class="`status-${item.status}`"></div>
          </div>
          <slot name="item" :item="item"></slot>
        </div>
      </router-link>
      <a class="nav-item" href="#" v-if="haveMore" @click="$emit('load-more')">
        <div class="nav-item-inner">
          <p>Load More</p>
        </div>
      </a>
    </div>
  </nav>
</template>

<script lang="ts">
  import Vue, {PropType} from 'vue';

  export default Vue.extend({
    name: 'TabNav',
    props: {
      title: {
        type: String,
        required: true,
      },
      stats: Object as PropType<{[name: string]: any}>,
      linkGen: {
        type: Function as PropType<(itemId: string) => string>,
        required: true,
      },
      items: Array as PropType<Array<any>>,
      haveMore: Boolean,
    },
  });
</script>

<style scoped>
  .nav {
    --padding: 10px;

    width: max-content;
    max-width: 25em;
    height: 100vh;

    display: flex;
    flex-direction: column;
  }

  .nav-header {
    padding: var(--padding);
  }

  .nav-header > *:not(:last-child) {
    margin-bottom: var(--padding);
  }

  .nav-header-title {
    font-size: 1rem;
    font-weight: 400;

    margin: 0;
  }

  .nav-header-stats {
    display: flex;
  }

  .nav-header-stats > *:not(:last-child) {
    margin-right: 1em;
  }

  .nav-items {
    overflow-y: scroll;
  }

  .nav-item {
    border-top: 1px solid var(--line-color);
    color: inherit;
    text-decoration: none;

    display: block;
  }

  a.nav-item:hover {
    background-color: var(--hover-color);
  }

  .nav-item-inner {
    border-left: 3px solid transparent;

    padding: var(--padding);
    padding-left: calc(var(--padding) - 3px);
    display: flex;
    align-items: center;
  }

  .nav-item-inner > *:not(.flex-spacer):not(:last-child) {
    margin-right: var(--padding)
  }

  .nav-item.router-link-active .nav-item-inner {
    border-left-color: var(--highlight-color);
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

    animation: spin var(--anim-speed) linear infinite;

    margin: 0;
  }

  .status-icon.status-passed {
    background-color: var(--status-passed-color);
  }

  .status-icon.status-failed, .status-icon.status-errored {
    background-color: var(--status-failed-color);
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>
