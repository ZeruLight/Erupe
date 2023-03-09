BEGIN;

ALTER TABLE public.characters
    ADD COLUMN mezfes BYTEA;

END;