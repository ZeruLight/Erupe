BEGIN;

ALTER TABLE characters ADD bonus_quests INT NOT NULL DEFAULT 0;

ALTER TABLE characters ADD daily_quests INT NOT NULL DEFAULT 0;

ALTER TABLE characters ADD promo_points INT NOT NULL DEFAULT 0;

END;