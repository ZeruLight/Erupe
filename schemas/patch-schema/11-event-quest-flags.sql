BEGIN;

ALTER TABLE IF EXISTS public.event_quests ADD COLUMN IF NOT EXISTS flags integer;

END;