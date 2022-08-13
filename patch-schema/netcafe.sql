BEGIN;

ALTER TABLE IF EXISTS public.characters
    ADD COLUMN cafe_time integer DEFAULT 0;

CREATE TABLE IF NOT EXISTS public.cafebonus
(
    id integer NOT NULL PRIMARY KEY,
    seconds_req integer NOT NULL,
    item_type integer NOT NULL,
    item_id integer NOT NULL,
    quantity integer NOT NULL
);

CREATE TABLE IF NOT EXISTS public.cafe_accepted
(
    cafe_id integer NOT NULL,
    character_id integer NOT NULL
);

END;