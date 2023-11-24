BEGIN;

ALTER TABLE IF EXISTS public.event_quests ADD COLUMN IF NOT EXISTS start_time timestamp with time zone NOT NULL DEFAULT (CURRENT_DATE + interval '0 second');
ALTER TABLE IF EXISTS public.event_quests ADD COLUMN IF NOT EXISTS active_duration int DEFAULT 4;
ALTER TABLE IF EXISTS public.event_quests ADD COLUMN IF NOT EXISTS inactive_duration int DEFAULT 3;

END;