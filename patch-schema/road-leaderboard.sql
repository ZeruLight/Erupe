BEGIN;

CREATE TABLE IF NOT EXISTS rengoku_score (
    character_id integer PRIMARY KEY,
    max_stages_mp integer,
    max_points_mp integer,
    max_stages_sp integer,
    max_points_sp integer
);

END;