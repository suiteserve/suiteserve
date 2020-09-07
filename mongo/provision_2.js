// noinspection JSUnresolvedVariable,JSUnresolvedFunction
db.createUser({
  user: 'admin',
  pwd: 'admin',
  roles: ['root'],
});

// noinspection JSUnresolvedVariable
db.auth('admin', 'admin');

// noinspection JSUnresolvedFunction,JSUnresolvedVariable
db.createUser({
  user: 'suiteserve',
  pwd: 'suiteserve',
  roles: [{role: 'readWrite', db: 'suiteserve'}],
});
