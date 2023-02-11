BEGIN;

ALTER TABLE IF EXISTS public.guild_characters ADD donated_rp INT DEFAULT 0;

END;