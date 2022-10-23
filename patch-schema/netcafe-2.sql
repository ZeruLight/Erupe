BEGIN;

ALTER TABLE IF EXISTS public.characters
    ADD COLUMN cafe_reset timestamp without time zone;

END;