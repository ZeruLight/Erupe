--adds world_name and land columns
BEGIN;

CREATE TABLE IF NOT EXISTS public.servers
(
    server_id integer NOT NULL,
    season integer NOT NULL,
    current_players integer NOT NULL,
    world_name text COLLATE pg_catalog."default",
    world_description text,
    land integer
);

ALTER TABLE public.servers
    ADD COLUMN IF NOT EXISTS land integer;

ALTER TABLE public.servers
    ADD COLUMN IF NOT EXISTS world_name text COLLATE pg_catalog."default";

ALTER TABLE public.servers
    ADD COLUMN IF NOT EXISTS world_description text;

END;