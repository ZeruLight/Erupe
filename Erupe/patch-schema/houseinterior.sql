BEGIN;

ALTER TABLE IF EXISTS public.characters
    ADD COLUMN IF NOT EXISTS house bytea;

END;