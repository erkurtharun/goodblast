CREATE TABLE users
(
    id            BIGSERIAL PRIMARY KEY,
    username      TEXT UNIQUE          NOT NULL,
    password_hash TEXT                 NOT NULL,
    coins         BIGINT  DEFAULT 1000 NOT NULL,
    level         INTEGER DEFAULT 1    NOT NULL,
    country       TEXT
);