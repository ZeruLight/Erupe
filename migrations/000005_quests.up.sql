BEGIN;

CREATE TABLE questlists (
  ind int NOT NULL PRIMARY KEY,
  questlist bytea
);

END;	


INSERT INTO questlists (id, questlist) VALUES ('0', pg_read_binary_file('c:\save\quest0.bin'));
INSERT INTO questlists (id, questlist) VALUES ('42', pg_read_binary_file('c:\save\quest1.bin'));
INSERT INTO questlists (id, questlist) VALUES ('84', pg_read_binary_file('c:\save\quest2.bin'));
INSERT INTO questlists (id, questlist) VALUES ('126', pg_read_binary_file('c:\save\quest3.bin'));
INSERT INTO questlists (id, questlist) VALUES ('168', pg_read_binary_file('c:\save\quest4.bin'));
INSERT INTO questlists (id, questlist) VALUES ('5', pg_read_binary_file('c:\save\quest5.bin'));

		ackHandle := bf.ReadUint32()
		bf.ReadBytes(5)
		questList := bf.ReadUint8()
		bf.ReadBytes(1)
		if questList == 0 {
			questListCount = 0
		} else if (questList <= 44)  {
			questListCount = 1
		} else if questList <= 88 {
			questListCount = 2
		} else if questList <= 126 {
			questListCount = 3
		} else if questList <= 172 {
			questListCount = 4
		} else {
			questListCount = 0