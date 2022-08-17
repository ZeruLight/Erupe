BEGIN;

CREATE SEQUENCE IF NOT EXISTS public.rasta_id_seq;

UPDATE characters SET savemercenary=NULL;

END;