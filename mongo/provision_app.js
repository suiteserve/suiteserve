// noinspection JSUnresolvedVariable,JSUnresolvedFunction
db.createUser({
  user: 'suiteserve',
  pwd: 'suiteserve',
  roles: [
    {role: 'readWrite', db: 'suiteserve'},
  ],
})
