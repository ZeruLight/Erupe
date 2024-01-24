BEGIN;

ALTER TABLE IF EXISTS public.users ADD COLUMN discord_token text;
ALTER TABLE IF EXISTS public.users ADD COLUMN discord_id text;

END;