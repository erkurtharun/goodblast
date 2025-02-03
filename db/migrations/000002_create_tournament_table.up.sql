CREATE TABLE tournaments
(
    id         BIGSERIAL PRIMARY KEY,
    start_date TIMESTAMP   NOT NULL,
    end_date   TIMESTAMP   NOT NULL,
    status     VARCHAR(16) NOT NULL DEFAULT 'planned'
);