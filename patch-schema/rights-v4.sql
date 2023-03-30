BEGIN;

-- Remove Trial Course from all users
UPDATE users SET rights = rights-2;

ALTER TABLE IF EXISTS public.users
    ALTER COLUMN rights SET DEFAULT 12;

END;