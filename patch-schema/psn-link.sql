BEGIN;

ALTER TABLE public.sign_sessions ADD COLUMN id SERIAL;

ALTER TABLE public.sign_sessions ADD CONSTRAINT sign_sessions_pkey PRIMARY KEY (id);

ALTER TABLE public.sign_sessions ALTER COLUMN user_id DROP NOT NULL;

ALTER TABLE public.sign_sessions ADD COLUMN psn_id TEXT;

END;