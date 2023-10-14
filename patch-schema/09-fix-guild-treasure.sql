BEGIN;

ALTER TABLE IF EXISTS public.guild_hunts DROP COLUMN IF EXISTS hunters;

ALTER TABLE IF EXISTS public.guild_characters
    ADD COLUMN treasure_hunt integer;

END;