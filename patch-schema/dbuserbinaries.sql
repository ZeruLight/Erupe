BEGIN;

CREATE TABLE user_binaries
(
    id int PRIMARY KEY,
    type1 bytea,
    type2 bytea,
    type3 bytea
);

END;