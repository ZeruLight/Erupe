BEGIN;

-- This will delete all of your old distribution data!
--ALTER TABLE IF EXISTS public.distribution DROP COLUMN IF EXISTS data;

CREATE TABLE public.distribution_items
(
    id serial PRIMARY KEY,
    distribution_id integer,
    item_type integer,
    item_id integer,
    quantity integer
);

END;