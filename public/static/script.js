const app = new Vue({
    el: '#app',
    data: {
        suites: [],
    },
    methods: {
        openSuite: function(event, suite) {
            const e = event.currentTarget;
            const statusIcon = e.querySelector('.status-icon');
            const statusIconClasses = statusIcon.classList;

            if (statusIconClasses.contains('created')) {
                statusIconClasses.replace('created', 'running');
            } else if (statusIconClasses.contains('running')) {
                statusIconClasses.replace('running', 'finished');
            } else {
                statusIconClasses.replace('finished', 'running');
            }

            fetch(`/suites/${suite.id}/cases`)
                .then(res => res.json()
                    .then(cases => suite.cases = cases));
        },
        formatTime: function(millis) {
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

(function () {
    fetch('/suites')
        .then(res => res.json()
            .then(suites => app.suites = suites));
})()