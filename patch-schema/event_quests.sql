BEGIN;

create table if not exists event_quests
(
    id serial primary key,
    max_players integer,
    quest_type integer not null,
    quest_id integer not null,
    mark integer
);

END;