BEGIN;

ALTER TABLE IF EXISTS public.normal_shop_items
    RENAME TO shop_items;

ALTER TABLE IF EXISTS public.shop_items
    RENAME COLUMN shoptype TO shop_type;

ALTER TABLE IF EXISTS public.shop_items
    RENAME COLUMN shopid TO shop_id;

ALTER TABLE IF EXISTS public.shop_items
    RENAME COLUMN itemhash TO id;

ALTER TABLE IF EXISTS public.shop_items
    RENAME COLUMN itemid TO item_id;

ALTER TABLE IF EXISTS public.shop_items
    ALTER COLUMN points TYPE integer;

ALTER TABLE IF EXISTS public.shop_items
    RENAME COLUMN points TO cost;

ALTER TABLE IF EXISTS public.shop_items
    RENAME COLUMN tradequantity TO quantity;

ALTER TABLE IF EXISTS public.shop_items
    RENAME COLUMN rankreqlow TO min_hr;

ALTER TABLE IF EXISTS public.shop_items
    RENAME COLUMN rankreqhigh TO min_sr;

ALTER TABLE IF EXISTS public.shop_items
    RENAME COLUMN rankreqg TO min_gr;

ALTER TABLE IF EXISTS public.shop_items
    RENAME COLUMN storelevelreq TO req_store_level;

ALTER TABLE IF EXISTS public.shop_items
    RENAME COLUMN maximumquantity TO max_quantity;

ALTER TABLE IF EXISTS public.shop_items
    DROP COLUMN IF EXISTS boughtquantity;

ALTER TABLE IF EXISTS public.shop_items
    RENAME COLUMN roadfloorsrequired TO road_floors;

ALTER TABLE IF EXISTS public.shop_items
    RENAME COLUMN weeklyfataliskills TO road_fatalis;

DROP TABLE IF EXISTS public.shop_item_state;

CREATE TABLE IF NOT EXISTS public.shop_items_bought (
    character_id INTEGER,
    shop_item_id INTEGER,
    bought INTEGER
);

END;