BEGIN;

ALTER TABLE IF EXISTS public.event_quests ADD COLUMN IF NOT EXISTS start_time timestamp with time zone NOT NULL DEFAULT now();
ALTER TABLE IF EXISTS public.event_quests ADD COLUMN IF NOT EXISTS active_duration int;
ALTER TABLE IF EXISTS public.event_quests ADD COLUMN IF NOT EXISTS inactive_duration int;
UPDATE public.event_quests SET active_duration=NULL, inactive_duration=NULL;
ALTER TABLE IF EXISTS public.event_quests RENAME active_duration TO active_days;
ALTER TABLE IF EXISTS public.event_quests RENAME inactive_duration TO inactive_days;

END;