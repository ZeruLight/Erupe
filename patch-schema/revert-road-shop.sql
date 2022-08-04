BEGIN;

ALTER TABLE IF EXISTS public.normal_shop_items
    DROP COLUMN IF EXISTS enable_weeks;

ALTER TABLE IF EXISTS public.shop_item_state
    DROP COLUMN IF EXISTS week;

END;