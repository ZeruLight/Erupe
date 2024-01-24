BEGIN;

ALTER TABLE IF EXISTS public.users ADD COLUMN op boolean;

CREATE TABLE public.bans
(
    user_id integer NOT NULL,
    expires timestamp with time zone,
    PRIMARY KEY (user_id)
);

END;