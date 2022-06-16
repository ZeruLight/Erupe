BEGIN;

ALTER TABLE IF EXISTS public.guilds
(
    ADD COLUMN pugi_name_1 character varying(12),
    ADD COLUMN pugi_name_2 character varying(12),
    ADD COLUMN pugi_name_3 character varying(12)
);

CREATE TABLE IF NOT EXISTS public.guild_alliances
(
	  id serial NOT NULL PRIMARY KEY,
    name character varying(24) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    parent_id int NOT NULL,
    sub1_id int,
    sub2_id int
);

END;