BEGIN;

create table if not exists event_quests
(
    id serial primary key,
    max_players integer,
    quest_type integer not null,
    quest_id integer not null,
    mark integer
);

ALTER TABLE IF EXISTS public.servers DROP COLUMN IF EXISTS season;

END;
