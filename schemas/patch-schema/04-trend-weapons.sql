CREATE TABLE public.trend_weapons
(
    weapon_id integer NOT NULL,
    weapon_type integer NOT NULL,
    count integer DEFAULT 0,
    PRIMARY KEY (weapon_id)
);