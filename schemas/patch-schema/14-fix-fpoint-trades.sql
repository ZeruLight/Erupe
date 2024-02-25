BEGIN;

DELETE FROM public.fpoint_items;
ALTER TABLE IF EXISTS public.fpoint_items ALTER COLUMN item_type SET NOT NULL;
ALTER TABLE IF EXISTS public.fpoint_items ALTER COLUMN item_id SET NOT NULL;
ALTER TABLE IF EXISTS public.fpoint_items ALTER COLUMN quantity SET NOT NULL;
ALTER TABLE IF EXISTS public.fpoint_items ALTER COLUMN fpoints SET NOT NULL;
ALTER TABLE IF EXISTS public.fpoint_items DROP COLUMN IF EXISTS trade_type;
ALTER TABLE IF EXISTS public.fpoint_items ADD COLUMN buyable boolean NOT NULL;

END;