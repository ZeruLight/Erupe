--adds world_name and land columns

CREATE TABLE IF NOT EXISTS public.servers
(
    server_id integer NOT NULL,
    season integer NOT NULL,
    current_players integer NOT NULL,
    world_name text COLLATE pg_catalog."default",
    land integer
)