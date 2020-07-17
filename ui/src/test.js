import * as pb from '@suiteserve/protocol-web';

const service = new pb.suiteServiceClient.SuiteServiceClient('https://localhost:8080');

const req = new pb.suiteServiceClient.CreateSuiteRequest();
req.setName('test1');
req.setPlannedCases(2);
req.setTagsList(['hello', 'world']);
const call = service.createSuite(req, {}, (err, resp) => {
  if (err) {
    console.log(err.code);
    console.log(err.message);
  } else {
    console.log(resp.getId());
  }
});
call.on('status', status => {
  console.log(status.code);
  console.log(status.details);
  console.log(status.metadata);
});
