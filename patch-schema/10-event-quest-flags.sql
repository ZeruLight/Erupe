BEGIN;

ALTER TABLE IF EXISTS public.event_quests ADD COLUMN IF NOT EXISTS flag_override integer NOT NULL DEFAULT -1;

END;