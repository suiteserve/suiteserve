

let activeSuiteElem;

const app = new Vue({
    el: '#app',
    created: function () {
        retry.bind(this)(() => true, fetchSuites)
            .then(suites => this.suites = suites);
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

            retry.bind(this)(() => true, fetchCases, suite)
                .then(cases => this.cases = cases)
                .catch(() => {});
        },
    },
});