<template>
  <TabNav title="Suites" :link-gen="genLink" :items="suites" :is-more="isMore" :stats="{
    'Running': runningCount,
    'Finished': finishedCount,
  }" @load-more="load">
    <template #item="{ item }">
      <div>
        <p>{{ item.name }}</p>
        <p class="muted">{{ formatUnix(item.started_at) }}</p>
      </div>
    </template>
  </TabNav>
</template>

<script>
  import TabNav from './TabNav';

  export default {
    name: 'Suites',
    data() {
      return {
        suites: [],
        runningCount: 0,
        finishedCount: 0,
        nextId: null,
      };
    },
    created() {
      this.load(true);
    },
    computed: {
      isMore() {
        return this.nextId !== null;
      },
    },
    methods: {
      genLink(suiteId) {
        return {
          name: 'suite',
          params: {
            suiteId,
          },
        };
      },
      formatUnix(millis) {
        return new Date(millis).toLocaleString(navigator.languages, {
          weekday: 'short',
          year: 'numeric',
          month: 'short',
          day: 'numeric',
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
        });
      },
      async load(init) {
        const url = new URL('/v1/suites', window.location.href);
        if (this.nextId) {
          url.searchParams.append('from_id', this.nextId);
        }
        url.searchParams.append('limit', '10');

        const res = await fetch(url.href);
        const json = await res.json();

        if (!res.ok) {
          throw `Error loading suites: ${json.error}`;
        }

        if (init) {
          this.suites = json.suites;
        } else {
          this.suites.push(...json.suites);
        }

        this.runningCount = json.running_count;
        this.finishedCount = json.finished_count;
        this.nextId = json.next_id;
      },
    },
    components: {
      TabNav,
    },
  };
</script>
