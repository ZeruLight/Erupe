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
    garden bytea
);

END;