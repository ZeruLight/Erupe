BEGIN;

ALTER TABLE IF EXISTS public.guild_characters ADD COLUMN trial_vote integer;

END;