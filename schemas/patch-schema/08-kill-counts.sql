CREATE TABLE public.kill_logs
(
    id serial,
    character_id integer NOT NULL,
    monster integer NOT NULL,
    quantity integer NOT NULL,
    timestamp timestamp with time zone NOT NULL,
    PRIMARY KEY (id)
);

ALTER TABLE IF EXISTS public.guild_characters
    ADD COLUMN box_claimed timestamp with time zone DEFAULT now();