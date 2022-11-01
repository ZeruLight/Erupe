BEGIN;

ALTER TABLE IF EXISTS public.characters
    ADD COLUMN IF NOT EXISTS cafe_time integer DEFAULT 0;

ALTER TABLE IF EXISTS public.characters
    DROP COLUMN IF EXISTS netcafe_points;

ALTER TABLE IF EXISTS public.characters
    ADD COLUMN IF NOT EXISTS netcafe_points int DEFAULT 0;

ALTER TABLE IF EXISTS public.characters
    ADD COLUMN IF NOT EXISTS boost_time timestamp without time zone;

CREATE TABLE IF NOT EXISTS public.cafebonus
(
    id serial NOT NULL PRIMARY KEY,
    time_req integer NOT NULL,
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