BEGIN;

ALTER TABLE IF EXISTS public.characters
    ADD COLUMN house bytea;

END;