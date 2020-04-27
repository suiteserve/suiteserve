const app = new Vue({
    el: '#app',
    data: {
        deleteAttachmentId: '',
    },
    methods: {
        createSuite: function (e) {
            document.getElementById("create-suite-form-created-at").value =
                Math.floor(new Date().getTime() / 1000);
        },
    },
});