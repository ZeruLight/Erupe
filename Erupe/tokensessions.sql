BEGIN;

DROP TABLE IF EXISTS public.sign_sessions;
CREATE TABLE IF NOT EXISTS public.sign_sessions
(
    user_id int NOT NULL,
    char_id int,
    token varchar(16) NOT NULL,
    server_id integer
);

DROP TABLE IF EXISTS public.servers;
CREATE TABLE IF NOT EXISTS public.servers
(
    server_id int NOT NULL,
    season int NOT NULL,
    current_players int NOT NULL
);

ALTER TABLE IF EXISTS public.characters
    ADD COLUMN deleted boolean NOT NULL DEFAULT false;

ALTER TABLE IF EXISTS public.characters
    ADD COLUMN friends text NOT NULL DEFAULT '';

END;
