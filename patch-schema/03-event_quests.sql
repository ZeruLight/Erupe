BEGIN;

create table if not exists event_quests
(
    id serial primary key,
    max_players integer,
    quest_type integer not null,
    quest_id integer not null,
    mark integer,
    weekly_cycle INT,
    available_in_all_cycles bool default true
);

ALTER TABLE IF EXISTS public.servers DROP COLUMN IF EXISTS season;

CREATE TABLE IF NOT EXISTS weekly_cycle_info (
    id SERIAL PRIMARY KEY,
    current_cycle_number INT,
    last_cycle_update_timestamp TIMESTAMP WITH TIME ZONE
);

END;
