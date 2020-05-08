// noinspection JSUnresolvedVariable,JSUnresolvedFunction
db.createUser({
    user: 'testpass',
    pwd: 'testpass',
    roles: [
        {role: 'readWrite', db: 'testpass'},
    ],
})
