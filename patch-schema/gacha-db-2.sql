BEGIN;

ALTER TABLE characters
    DROP COLUMN IF EXISTS gacha_prem;

ALTER TABLE characters
    DROP COLUMN IF EXISTS gacha_trial;

ALTER TABLE characters
    DROP COLUMN IF EXISTS frontier_points;

ALTER TABLE users
    ADD IF NOT EXISTS gacha_premium INT;

ALTER TABLE users
    ADD IF NOT EXISTS gacha_trial INT;

ALTER TABLE users
    ADD IF NOT EXISTS frontier_points INT;

DROP TABLE IF EXISTS public.gacha_shop;

CREATE TABLE IF NOT EXISTS public.gacha_shop (
    id SERIAL PRIMARY KEY,
    min_gr INTEGER,
    min_hr INTEGER,
    name TEXT,
    url_banner TEXT,
    url_feature TEXT,
    url_thumbnail TEXT,
    wide BOOLEAN,
    recommended BOOLEAN,
    gacha_type INTEGER,
    hidden BOOLEAN
);

DROP TABLE IF EXISTS public.gacha_shop_items;

CREATE TABLE IF NOT EXISTS public.gacha_entries (
    id            SERIAL PRIMARY KEY,
    gacha_id      INTEGER,
    entry_type    INTEGER,
    item_type     INTEGER,
    item_number   INTEGER,
    item_quantity INTEGER,
    weight        INTEGER,
    rarity        INTEGER,
    rolls         INTEGER,
    daily_limit   INTEGER
);

CREATE TABLE IF NOT EXISTS public.gacha_items (
    id        SERIAL PRIMARY KEY,
    entry_id  INTEGER,
    item_type INTEGER,
    item_id   INTEGER,
    quantity  INTEGER
);

END;