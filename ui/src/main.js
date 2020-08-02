import Vue from 'vue';
import App from '@/App';
import router from '@/router';
import * as rpc from '@suiteserve/protocol-web';

// Vue.config.productionTip = false;

const client = new rpc.QueryServiceClient('https://localhost:8080')
const req = new rpc.WatchSuitesRequest()
const stream = client.watchSuites(req);
stream.on('data', resp => {
  console.log(resp);
})
stream.on('end', () => {
  console.log('end');
})

new Vue({
  el: '#app',
  router,
  render: h => h(App),
});
