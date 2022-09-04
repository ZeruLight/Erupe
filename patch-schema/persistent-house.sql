BEGIN;

CREATE TABLE IF NOT EXISTS public.user_binary
(
    id serial NOT NULL PRIMARY KEY,
    type2 bytea,
    type3 bytea,
    house_tier bytea,
    house_state int,
    house_password text,
    house_data bytea,
    house_furniture bytea,
    bookshelf bytea,
    gallery bytea,
    tore bytea,
    garden bytea,
    mission bytea
);

-- Create entries for existing users
INSERT INTO public.user_binary (id) SELECT c.id FROM characters c;

-- Copy existing data
UPDATE public.user_binary
    SET house_furniture = (SELECT house FROM characters WHERE user_binary.id = characters.id);

UPDATE public.user_binary
    SET mission = (SELECT trophy FROM characters WHERE user_binary.id = characters.id);

-- Drop old data location
ALTER TABLE public.characters
    DROP COLUMN house;

ALTER TABLE public.characters
    DROP COLUMN trophy;

END;