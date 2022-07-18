BEGIN;

ALTER TABLE IF EXISTS public.guilds
    ADD COLUMN pugi_name_1 varchar(12) DEFAULT '';
ALTER TABLE IF EXISTS public.guilds
    ADD COLUMN pugi_name_2 varchar(12) DEFAULT '';
ALTER TABLE IF EXISTS public.guilds
    ADD COLUMN pugi_name_3 varchar(12) DEFAULT '';

CREATE TABLE IF NOT EXISTS public.guild_alliances
(
	  id serial NOT NULL PRIMARY KEY,
    name varchar(24) NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now(),
    parent_id int NOT NULL,
    sub1_id int,
    sub2_id int
);

CREATE TABLE IF NOT EXISTS public.guild_adventures
(
    id serial NOT NULL PRIMARY KEY,
    guild_id int NOT NULL,
    destination int NOT NULL,
    charge int NOT NULL DEFAULT 0,
    depart int NOT NULL,
    return int NOT NULL,
    collected_by text NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS public.guild_meals
(
	  id serial NOT NULL PRIMARY KEY,
    guild_id int NOT NULL,
    meal_id int NOT NULL,
    level int NOT NULL,
    expires int NOT NULL
);

END;