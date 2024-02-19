BEGIN;

CREATE TABLE festa_submissions (
    character_id int NOT NULL,
    guild_id int NOT NULL,
    trial_type int NOT NULL,
    souls int NOT NULL,
    timestamp timestamp with time zone NOT NULL
);

ALTER TABLE guild_characters DROP COLUMN souls;

ALTER TYPE festival_colour RENAME TO festival_color;

END;