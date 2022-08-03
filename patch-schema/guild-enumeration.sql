BEGIN;

ALTER TABLE IF EXISTS public.guilds
    ADD COLUMN IF NOT EXISTS recruiting bool NOT NULL DEFAULT true;

ALTER TABLE IF EXISTS public.guild_characters
    ADD COLUMN IF NOT EXISTS recruiter bool NOT NULL DEFAULT false;

END;