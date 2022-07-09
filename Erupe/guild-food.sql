BEGIN;

CREATE TABLE IF NOT EXISTS public.guild_meals
(
	  id serial NOT NULL PRIMARY KEY,
    guild_id int NOT NULL,
    meal_id int NOT NULL,
    level int NOT NULL,
    expires int NOT NULL
);

END;