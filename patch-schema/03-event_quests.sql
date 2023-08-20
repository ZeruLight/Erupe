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

ALTER TABLE IF EXISTS public.events ADD COLUMN IF NOT EXISTS current_cycle_number int;
ALTER TYPE event_type ADD VALUE 'EventQuests';

END;
