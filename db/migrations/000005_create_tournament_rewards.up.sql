CREATE TABLE tournament_rewards
(
    id            BIGSERIAL PRIMARY KEY,
    tournament_id BIGINT    NOT NULL REFERENCES tournaments (id) ON DELETE SET NULL,
    user_id       BIGINT    NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    rank          INT       NOT NULL,
    reward_coins  INT       NOT NULL,
    claimed       BOOLEAN   NOT NULL DEFAULT false,
    created_at    TIMESTAMP NOT NULL DEFAULT now()
);