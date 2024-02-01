BEGIN;

ALTER TABLE IF EXISTS public.guild_hunts DROP COLUMN IF EXISTS hunters;

ALTER TABLE IF EXISTS public.guild_characters
    ADD COLUMN treasure_hunt integer;

ALTER TABLE IF EXISTS public.guild_hunts
    ADD COLUMN start timestamp with time zone NOT NULL DEFAULT now();

UPDATE guild_hunts SET start=to_timestamp(return);

ALTER TABLE IF EXISTS public.guild_hunts DROP COLUMN IF EXISTS "return";

ALTER TABLE IF EXISTS public.guild_hunts
    RENAME claimed TO collected;

CREATE TABLE public.guild_hunts_claimed
(
    hunt_id integer NOT NULL,
    character_id integer NOT NULL
);

ALTER TABLE IF EXISTS public.guild_hunts DROP COLUMN IF EXISTS treasure;

END;