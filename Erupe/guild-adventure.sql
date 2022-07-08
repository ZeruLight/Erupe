BEGIN;

CREATE TABLE IF NOT EXISTS public.guild_adventures
(
    id serial NOT NULL PRIMARY KEY,
    guild_id int NOT NULL,
    destination int NOT NULL,
    charge int NOT NULL DEFAULT 0,
    depart int NOT NULL,
    return int NOT NULL,
    collected_by text NOT NULL DEFAULT ''
);

END;