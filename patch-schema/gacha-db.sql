BEGIN;

DROP TABLE IF EXISTS public.gacha_shop;

CREATE TABLE IF NOT EXISTS public.gacha_shop(
    id serial PRIMARY KEY,
    min_gr integer,
    min_hr integer,
    name text,
    link1 text,
    link2 text,
    link3 text,
    icon integer,
    type integer,
    hide boolean
);

END;