BEGIN;

INSERT INTO public.shop_items
(shop_type, shop_id, item_id, cost, quantity, min_hr, min_sr, min_gr, store_level, max_quantity, road_floors, road_fatalis)
VALUES
    (5,5,16516,100,1,0,0,1,0,0,0,0),
    (5,5,16517,100,1,0,0,1,0,0,0,0),
    (7,0,13190,10,1,0,0,0,0,0,0,0),
    (7,0,1662,10,1,0,0,0,0,0,0,0),
    (7,0,10179,100,1,0,0,0,0,0,0,0);

END;