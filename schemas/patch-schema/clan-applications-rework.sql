BEGIN;

ALTER TABLE IF EXISTS public.guild_applications DROP COLUMN IF EXISTS actor_id;

ALTER TABLE IF EXISTS public.guild_applications DROP COLUMN IF EXISTS application_type;

ALTER TABLE IF EXISTS public.guild_applications
    ALTER COLUMN created_at DROP NOT NULL;
ALTER TABLE IF EXISTS public.guild_applications DROP CONSTRAINT IF EXISTS guild_applications_actor_id_fkey;

create table public.guild_invites (
    id serial primary key,
    guild_id integer,
    character_id integer,
    actor_id integer,
    created_at timestamp with time zone not null default now(),
    foreign key (guild_id) references guilds (id),
    foreign key (character_id) references characters (id),
    foreign key (actor_id) references characters (id)
);

drop type if exists guild_application_type;

ALTER TABLE IF EXISTS public.mail DROP CONSTRAINT IF EXISTS mail_sender_id_fkey;

END;