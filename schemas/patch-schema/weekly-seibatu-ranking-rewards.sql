BEGIN;

CREATE TABLE IF NOT EXISTS weekly_seibatu_ranking_reward (
    id serial PRIMARY KEY,
    reward_id integer,
    index0 integer,
    index1 integer,
    index2 integer,
    distribution_type integer,
    item_id integer,
    amount integer
);

END;