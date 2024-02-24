BEGIN;

CREATE TABLE IF NOT EXISTS scenario_counter (
    id serial primary key,
    scenario_id numeric not null,
    category_id numeric not null
);

END;