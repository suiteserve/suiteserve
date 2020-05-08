// noinspection JSUnresolvedVariable,JSUnresolvedFunction
db.createUser({
    user: 'admin',
    pwd: 'admin',
    roles: [
        'root',
    ],
});
