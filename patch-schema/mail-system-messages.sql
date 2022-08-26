BEGIN;

ALTER TABLE IF EXISTS public.mail
    ADD COLUMN IF NOT EXISTS is_sys_message bool DEFAULT false;

UPDATE mail SET is_sys_message=false;

ALTER TABLE IF EXISTS public.mail
    DROP CONSTRAINT IF EXISTS mail_sender_id_fkey;

INSERT INTO public.characters (id, name) VALUES (0, '');

END;