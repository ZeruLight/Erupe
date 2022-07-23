BEGIN;
CREATE TABLE IF NOT EXISTS public.normal_shop_items
(
	shoptype integer,
    shopid integer,
    itemhash integer not null,
    itemid integer,
    points integer,
    tradequantity integer,
    rankreqlow integer,
    rankreqhigh integer,
    rankreqg integer,
    storelevelreq integer,
    maximumquantity integer,
    boughtquantity integer,
    roadfloorsrequired integer,
    weeklyfataliskills integer,
    enable_weeks character varying(8)
);

ALTER TABLE IF EXISTS public.normal_shop_items
    ADD COLUMN IF NOT EXISTS enable_weeks character varying(8);

CREATE TABLE IF NOT EXISTS public.shop_item_state
(
    char_id bigint REFERENCES characters (id),
	itemhash int UNIQUE NOT NULL,
	usedquantity int,
    week int
);

ALTER TABLE IF EXISTS public.shop_item_state
    ADD COLUMN IF NOT EXISTS week int;

END;