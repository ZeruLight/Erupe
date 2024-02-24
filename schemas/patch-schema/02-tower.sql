BEGIN;

CREATE TABLE IF NOT EXISTS tower (
    char_id INT,
    tr INT,
    trp INT,
    tsp INT,
    block1 INT,
    block2 INT,
    skills TEXT,
    gems TEXT
);

ALTER TABLE IF EXISTS guild_characters
    ADD COLUMN IF NOT EXISTS tower_mission_1 INT;

ALTER TABLE IF EXISTS guild_characters
    ADD COLUMN IF NOT EXISTS tower_mission_2 INT;

ALTER TABLE IF EXISTS guild_characters
    ADD COLUMN IF NOT EXISTS tower_mission_3 INT;

ALTER TABLE IF EXISTS guilds
    ADD COLUMN IF NOT EXISTS tower_mission_page INT DEFAULT 1;

ALTER TABLE IF EXISTS guilds
    ADD COLUMN IF NOT EXISTS tower_rp INT DEFAULT 0;

END;