BEGIN;

ALTER TABLE IF EXISTS public.characters
    ADD COLUMN scenariodata bytea;

END;