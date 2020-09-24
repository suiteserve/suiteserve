db.createUser({
  user: 'root',
  pwd: 'pass',
  roles: ['root'],
});

db.auth('root', 'pass');
db = db.getSiblingDB('suiteserve');

db.createUser({
  user: 'ssmigrate',
  pwd: 'pass',
  roles: [{role: 'dbOwner', db: 'suiteserve'}],
});
db.createUser({
  user: 'suiteserve',
  pwd: 'pass',
  roles: [{role: 'readWrite', db: 'suiteserve'}],
});
