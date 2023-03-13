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
    RENAME COLUMN storelevelreq TO store_level;

ALTER TABLE IF EXISTS public.shop_items
    RENAME COLUMN maximumquantity TO max_quantity;

ALTER TABLE IF EXISTS public.shop_items
    DROP COLUMN IF EXISTS boughtquantity;

ALTER TABLE IF EXISTS public.shop_items
    RENAME COLUMN roadfloorsrequired TO road_floors;

ALTER TABLE IF EXISTS public.shop_items
    RENAME COLUMN weeklyfataliskills TO road_fatalis;

ALTER TABLE public.shop_items
    RENAME CONSTRAINT normal_shop_items_pkey TO shop_items_pkey;

ALTER TABLE IF EXISTS public.shop_items
    DROP CONSTRAINT IF EXISTS normal_shop_items_itemhash_key;

CREATE SEQUENCE public.shop_items_id_seq;

ALTER SEQUENCE public.shop_items_id_seq
    OWNER TO postgres;

ALTER TABLE IF EXISTS public.shop_items
    ALTER COLUMN id SET DEFAULT nextval('shop_items_id_seq'::regclass);

DROP TABLE IF EXISTS public.shop_item_state;

CREATE TABLE IF NOT EXISTS public.shop_items_bought (
    character_id INTEGER,
    shop_item_id INTEGER,
    bought INTEGER
);

END;