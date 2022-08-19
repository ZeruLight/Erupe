BEGIN;

CREATE TABLE IF NOT EXISTS public.stamps (
    character_id integer PRIMARY KEY,
    hl_total integer DEFAULT 0,
    hl_redeemed integer DEFAULT 0,
    hl_next timestamp without time zone,
    ex_total integer DEFAULT 0,
    ex_redeemed integer DEFAULT 0,
    ex_next timestamp without time zone
);

END;