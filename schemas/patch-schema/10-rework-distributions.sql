BEGIN;

-- This will delete all of your old distribution data!
--ALTER TABLE IF EXISTS public.distribution DROP COLUMN IF EXISTS data;

CREATE TABLE public.distribution_items
(
    id serial PRIMARY KEY,
    distribution_id integer NOT NULL,
    item_type integer NOT NULL,
    item_id integer,
    quantity integer
);

ALTER TABLE IF EXISTS public.distribution ALTER COLUMN min_hr DROP DEFAULT;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN max_hr DROP DEFAULT;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN min_sr DROP DEFAULT;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN max_sr DROP DEFAULT;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN min_gr DROP DEFAULT;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN max_gr DROP DEFAULT;

ALTER TABLE IF EXISTS public.distribution ALTER COLUMN min_hr DROP NOT NULL;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN max_hr DROP NOT NULL;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN min_sr DROP NOT NULL;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN max_sr DROP NOT NULL;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN min_gr DROP NOT NULL;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN max_gr DROP NOT NULL;

UPDATE distribution SET min_hr=NULL WHERE min_hr=65535;
UPDATE distribution SET max_hr=NULL WHERE max_hr=65535;
UPDATE distribution SET min_sr=NULL WHERE min_sr=65535;
UPDATE distribution SET max_sr=NULL WHERE max_sr=65535;
UPDATE distribution SET min_gr=NULL WHERE min_gr=65535;
UPDATE distribution SET max_gr=NULL WHERE max_gr=65535;

END;