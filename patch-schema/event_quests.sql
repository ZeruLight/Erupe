BEGIN;

create table if not exists event_quests
(
    id          serial primary key,
    max_players integer,
    quest_type  integer,
    quest_id    uint16 not null
);

END;