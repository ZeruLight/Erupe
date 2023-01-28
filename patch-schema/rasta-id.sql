BEGIN;

UPDATE characters SET savemercenary = NULL;

ALTER TABLE characters ADD rasta_id INT;

ALTER TABLE characters ADD pact_id INT;

END;