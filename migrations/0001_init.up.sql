CREATE TABLE IF NOT EXISTS users (
    uuid varchar(36) NOT NULL PRIMARY KEY,
    name text NOT NULL,
    email text NOT NULL,
    password text NOT NULL,
    UNIQUE (email)
);

CREATE TABLE IF NOT EXISTS tasks (
    id varchar(36) NOT NULL PRIMARY KEY,
    userid varchar(36) NOT NULL,
    status text NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    deleted boolean NOT NULL DEFAULT false
);