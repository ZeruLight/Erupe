BEGIN;

ALTER TABLE public.characters ADD COLUMN IF NOT EXISTS conquest_data BYTEA;

END;