CREATE TABLE users (
    id integer primary key AUTOINCREMENT,
    username text,
    firstname text,
    lastname text,
    email text,
    password char[40],
    salt char[16],
    role int
);

CREATE UNIQUE INDEX email on users(email);
CREATE UNIQUE INDEX name on users(username);
