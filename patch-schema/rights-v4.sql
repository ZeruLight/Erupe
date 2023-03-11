BEGIN;

-- Remove Trial Course from all users
UPDATE users SET rights = rights-2;

END;