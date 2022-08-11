BEGIN;

CREATE TABLE IF NOT EXISTS public.titles
(
    id int NOT NULL,
    char_id int NOT NULL,
    unlocked_at timestamp without time zone,
    updated_at timestamp without time zone
);

END;