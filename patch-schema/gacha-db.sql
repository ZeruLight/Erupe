BEGIN;

DROP TABLE IF EXISTS public.gacha_shop;

CREATE TABLE IF NOT EXISTS public.gacha_shop (
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

DROP TABLE IF EXISTS public.fpoint_items;

CREATE TABLE IF NOT EXISTS public.fpoint_items (
    id serial PRIMARY KEY,
    item_type integer,
    item_id integer,
    quantity integer,
    fpoints integer,
    trade_type integer
);

END;