let activeSuiteElem;

const app = new Vue({
    el: '#app',
    created: function () {
        fetch('/suites')
            .then(res => res.json()
                .then(suites => this.suites = suites));
    },
    data: {
        suites: [],
        cases: [],
        runningSuites: 0,
    },
    watch: {
        suites: {
            deep: true,
            handler: function (suites) {
                this.runningSuites = suites
                    .filter(s => s.status === 'running')
                    .length;
            },
        }
    },
    methods: {
        openSuite: function (event, suite) {
            event.preventDefault();
            const e = event.currentTarget;

            if (activeSuiteElem) {
                activeSuiteElem.classList.remove('active');
            }
            activeSuiteElem = e;
            activeSuiteElem.classList.add('active');

            const statusIcon = e.querySelector('.status-icon');
            const statusIconClasses = statusIcon.classList;

            if (statusIconClasses.contains('created')) {
                statusIconClasses.replace('created', 'running');
                suite.status = 'running';
            } else if (statusIconClasses.contains('running')) {
                statusIconClasses.replace('running', 'finished');
                suite.status = 'finished';
            } else {
                statusIconClasses.replace('finished', 'running');
                suite.status = 'running';
            }

            fetch(`/suites/${suite.id}/cases`)
                .then(res => res.json()
                    .then(cases => this.cases = cases));
        },
        formatTime: function (millis) {
            const date = new Date(millis);
            const opts = {
                weekday: 'short',
                year: 'numeric',
                month: 'short',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit',
            };
            return date.toLocaleString(navigator.languages, opts);
        }
    },
});