BEGIN;

DROP TABLE IF EXISTS public.login_boost_state;

CREATE TABLE IF NOT EXISTS public.login_boost (
    char_id INTEGER,
    week_req INTEGER,
    expiration TIMESTAMP WITH TIME ZONE,
    reset TIMESTAMP WITH TIME ZONE
);

END;