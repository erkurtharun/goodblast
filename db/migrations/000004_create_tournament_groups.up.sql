CREATE TABLE groups
(
    id            BIGSERIAL PRIMARY KEY,
    tournament_id BIGINT NOT NULL REFERENCES tournaments (id) ON DELETE CASCADE,
    group_number  INT    NOT NULL,
    current_size  INT    NOT NULL DEFAULT 0,

    UNIQUE (tournament_id, group_number)
);