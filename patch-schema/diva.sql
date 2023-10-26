BEGIN;

ALTER TABLE IF EXISTS public.guilds
    ADD COLUMN IF NOT EXISTS interception_maps bytea;

ALTER TABLE IF EXISTS public.guild_characters
    ADD COLUMN IF NOT EXISTS interception_points bytea;

CREATE TABLE IF NOT EXISTS public.diva_prizes (
    id SERIAL PRIMARY KEY,
    type PRIZE_TYPE,
    points_req INTEGER,
    item_type INTEGER,
    item_id INTEGER,
    quantity INTEGER,
    gr BOOLEAN,
    repeatable BOOLEAN
);

CREATE TABLE IF NOT EXISTS public.diva_beads (
    type INTEGER
);

ALTER TABLE IF EXISTS public.characters
    ADD COLUMN IF NOT EXISTS diva_bead INTEGER;

CREATE TABLE IF NOT EXISTS public.diva_beads_points (
    character_id INTEGER,
    points INTEGER,
    timestamp TIMESTAMP WITH TIME ZONE,
    bead_index INTEGER
);

END;