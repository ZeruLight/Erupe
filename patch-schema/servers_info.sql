--adds world_name and land columns
BEGIN;

CREATE TABLE IF NOT EXISTS public.servers
(
    server_id integer NOT NULL,
    season integer NOT NULL,
    current_players integer NOT NULL,
    world_name text COLLATE pg_catalog."default",
    land integer
)


ALTER TABLE IF EXISTS public.servers
    ADD COLUMN land integer;

ALTER TABLE IF EXISTS public.servers
    ADD COLUMN world_name text COLLATE pg_catalog."default";

END;