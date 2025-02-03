CREATE TABLE tournament_users
(
    id            BIGSERIAL PRIMARY KEY,
    tournament_id BIGINT    NOT NULL REFERENCES tournaments (id) ON DELETE CASCADE,
    user_id       BIGINT    NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    group_id      BIGINT    NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
    score         INT                DEFAULT 0,
    created_at    TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE (tournament_id, user_id)
);