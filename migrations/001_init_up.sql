CREATE TABLE links (
    id          SERIAL PRIMARY KEY,
    short_code  VARCHAR(10) NOT NULL UNIQUE,
    original_url TEXT       NOT NULL,
    created_at  TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE TABLE clicks (
    id          BIGSERIAL PRIMARY KEY,
    link_id     INTEGER     NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    user_agent  TEXT        NOT NULL DEFAULT '',
    ip          VARCHAR(45) NOT NULL DEFAULT '',
    clicked_at  TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_clicks_link_id_clicked_at ON clicks(link_id, clicked_at);