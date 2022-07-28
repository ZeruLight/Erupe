BEGIN;

ALTER TABLE IF EXISTS public.users
    ALTER rights SET DEFAULT 14;

ALTER TABLE IF EXISTS public.users
    ALTER rights SET NOT NULL;

UPDATE public.users SET rights=14 WHERE rights IS NULL;

END;