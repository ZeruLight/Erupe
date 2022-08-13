BEGIN;

ALTER TABLE IF EXISTS public.characters
    ADD COLUMN cafe_time integer DEFAULT 0;

CREATE TABLE IF NOT EXISTS public.cafebonus
(
    id integer NOT NULL PRIMARY KEY,
    line integer NOT NULL,
    itemclass integer NOT NULL,
    itemid integer NOT NULL,
    tradequantity integer NOT NULL
);

CREATE TABLE IF NOT EXISTS public.cafe_accepted
(
    cafe_id integer NOT NULL,
    character_id integer NOT NULL
);

END;