BEGIN;

CREATE TABLE IF NOT EXISTS public.feature_weapon
(
    start_time timestamp without time zone NOT NULL,
    featured integer NOT NULL
);

END;