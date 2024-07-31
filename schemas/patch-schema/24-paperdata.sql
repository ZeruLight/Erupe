BEGIN;

CREATE TABLE IF NOT EXISTS paper_data_gifts (
   id serial PRIMARY KEY,
   gift_id integer, 
   item_id integer, 
   unk0 integer,
   unk1 integer,
   chance integer
);

CREATE TABLE IF NOT EXISTS paper_data (
   id serial PRIMARY KEY,
   paper_type integer,
   paper_id integer, 
   option1 integer, 
   option2 integer,
   option3 integer,
   option4 integer,
   option5 integer
);

END;