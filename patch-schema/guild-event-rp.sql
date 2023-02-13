BEGIN;

ALTER TABLE IF EXISTS public.guild_characters ADD rp_today INT DEFAULT 0;

ALTER TABLE IF EXISTS public.guild_characters ADD rp_yesterday INT DEFAULT 0;

END;