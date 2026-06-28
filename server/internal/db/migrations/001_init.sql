CREATE TABLE IF NOT EXISTS players (
    username   TEXT PRIMARY KEY CHECK (username ~ '^[a-z]+$'),
    wins       INT NOT NULL DEFAULT 0,
    losses     INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS matches (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    seed        BIGINT NOT NULL,
    red_score   INT NOT NULL,
    green_score INT NOT NULL,
    winner      TEXT NOT NULL CHECK (winner IN ('RED', 'GREEN', 'DRAW')),
    duration_ms INT NOT NULL,
    ended_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS match_players (
    match_id BIGINT NOT NULL REFERENCES matches (id) ON DELETE CASCADE,
    username TEXT NOT NULL REFERENCES players (username),
    team     TEXT NOT NULL CHECK (team IN ('RED', 'GREEN')),
    result   TEXT NOT NULL CHECK (result IN ('win', 'loss', 'draw')),
    PRIMARY KEY (match_id, username)
);

CREATE INDEX IF NOT EXISTS idx_players_wins ON players (wins DESC, losses ASC);
