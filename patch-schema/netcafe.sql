BEGIN;

ALTER TABLE IF EXISTS public.characters
    ADD COLUMN cafe_time integer DEFAULT 0;

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

INSERT INTO public.cafebonus (time_req, item_type, item_id, quantity)
VALUES
    (1800, 17, 0, 250),
    (3600, 17, 0, 500),
    (7200, 17, 0, 1000),
    (10800, 17, 0, 1500),
    (18000, 17, 0, 1750),
    (28800, 17, 0, 3000),
    (43200, 17, 0, 4000);

END;