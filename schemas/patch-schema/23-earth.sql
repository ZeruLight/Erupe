BEGIN;

-- Add 'earth' to the event_type ENUM type
ALTER TYPE event_type ADD VALUE 'earth';

END;