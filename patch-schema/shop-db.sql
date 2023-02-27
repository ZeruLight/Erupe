BEGIN;

DROP TABLE IF EXISTS public.normal_shop_items;

CREATE TABLE IF NOT EXISTS public.shop_items (
    id SERIAL PRIMARY KEY,
    shop_type INTEGER,
    shop_id INTEGER,
    item_id INTEGER,
    cost INTEGER,
    quantity INTEGER,
    min_hr INTEGER,
    min_sr INTEGER,
    min_gr INTEGER,
    store_level INTEGER,
    max_quantity INTEGER,
    road_floors INTEGER,
    road_fatalis INTEGER
);

DROP TABLE IF EXISTS public.shop_item_state;

CREATE TABLE IF NOT EXISTS public.shop_items_bought (
    character_id INTEGER,
    shop_item_id INTEGER,
    bought INTEGER
);

END;