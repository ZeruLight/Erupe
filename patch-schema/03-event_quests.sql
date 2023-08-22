BEGIN;

create table if not exists event_quests
(
    id serial primary key,
    max_players integer,
    quest_type integer not null,
    quest_id integer not null,
    mark integer,
    start_time timestamp with time zone NOT NULL DEFAULT (CURRENT_DATE + interval '0 second'),
    active_duration int not null,
    inactive_duration int not null
);

ALTER TABLE IF EXISTS public.servers DROP COLUMN IF EXISTS season;

END;
