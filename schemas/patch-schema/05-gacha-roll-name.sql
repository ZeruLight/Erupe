BEGIN;

ALTER TABLE IF EXISTS public.gacha_entries
    ADD COLUMN name text;

END;