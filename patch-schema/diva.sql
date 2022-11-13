BEGIN;

ALTER TABLE IF EXISTS public.guilds
    ADD COLUMN IF NOT EXISTS interception_maps bytea;

END;