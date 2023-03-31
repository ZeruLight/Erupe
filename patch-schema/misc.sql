BEGIN;

CREATE TABLE IF NOT EXISTS public.feature_weapon
(
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    featured INTEGER NOT NULL
);

DROP TABLE IF EXISTS public.user_binaries;

DROP PROCEDURE IF EXISTS raviinit;

DROP PROCEDURE IF EXISTS ravireset;

END;