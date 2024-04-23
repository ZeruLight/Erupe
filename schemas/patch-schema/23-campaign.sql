BEGIN;

CREATE TABLE IF NOT EXISTS public.campaigns (
  id SERIAL PRIMARY KEY,
  min_hr INTEGER,
  max_hr INTEGER,
  min_sr INTEGER,
  max_sr INTEGER,
  min_gr INTEGER,
  max_gr INTEGER,
  recieve_type INTEGER,
  stamp_amount INTEGER,
  hide INTEGER,
  background_id INTEGER,
  hide_npc BOOLEAN,
  start_time TIMESTAMP WITH TIME ZONE,
  end_time TIMESTAMP WITH TIME ZONE,
  period_ended BOOLEAN,
  string0 TEXT,
  string1 TEXT,
  string2 TEXT,
  string3 TEXT,
  link TEXT,
  code_prefix TEXT
);

CREATE TABLE IF NOT EXISTS public.campaign_categories (
  id SERIAL PRIMARY KEY,
  cat_type INTEGER,
  title TEXT,
  description_text TEXT
);

CREATE TABLE IF NOT EXISTS public.campaign_category_links (
  id SERIAL PRIMARY KEY,
  campaign_id INTEGER,
  category_id INTEGER
);

CREATE TABLE IF NOT EXISTS public.campaign_entries (
  id SERIAL PRIMARY KEY,
  campaign_id INTEGER,
  hide BOOLEAN,
  item_type INTEGER,
  item_amount INTEGER,
  item_no INTEGER,
  unk1 INTEGER,
  unk2 INTEGER,
  deadline TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS public.campaign_state (
  id SERIAL PRIMARY KEY,
  campaign_id INTEGER,
  character_id INTEGER,
  code TEXT
);

END;