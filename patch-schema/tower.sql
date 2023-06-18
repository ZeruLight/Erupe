BEGIN;

CREATE TABLE IF NOT EXISTS tower (
    char_id INT,
    tr INT,
    trp INT,
    tsp INT,
    block1 INT,
    block2 INT,
    skills TEXT DEFAULT '0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0',
    gems TEXT DEFAULT '0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0'
);

END;