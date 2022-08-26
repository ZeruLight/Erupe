BEGIN;

ALTER TABLE IF EXISTS public.users
    ADD COLUMN IF NOT EXISTS last_login timestamp without time zone;

ALTER TABLE IF EXISTS public.users
    ADD COLUMN IF NOT EXISTS return_expires timestamp without time zone;

END;