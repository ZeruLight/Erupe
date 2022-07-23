BEGIN;

ALTER TABLE IF EXISTS public.characters
    ADD COLUMN IF NOT EXISTS scenariodata bytea;

ALTER TABLE IF EXISTS public.characters
    ADD COLUMN IF NOT EXISTS savefavoritequest bytea;

END;