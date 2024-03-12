BEGIN;

CREATE TABLE public.screenshots
(
    id serial PRIMARY KEY, 
    article_id TEXT NOT NULL,
    discord_message_id TEXT,
    char_id integer NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    discord_img_url TEXT,    
    );
END;