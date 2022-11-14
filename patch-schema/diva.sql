BEGIN;

ALTER TABLE IF EXISTS public.guilds
    ADD COLUMN IF NOT EXISTS interception_maps bytea;

ALTER TABLE IF EXISTS public.guild_characters
    ADD COLUMN IF NOT EXISTS interception_points bytea;

END;