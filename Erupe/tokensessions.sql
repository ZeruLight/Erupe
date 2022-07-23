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
    ADD COLUMN IF NOT EXISTS deleted boolean NOT NULL DEFAULT false;

ALTER TABLE IF EXISTS public.characters
    ADD COLUMN IF NOT EXISTS friends text NOT NULL DEFAULT '';

ALTER TABLE IF EXISTS public.users
    ADD COLUMN IF NOT EXISTS last_character int DEFAULT 0;

END;
