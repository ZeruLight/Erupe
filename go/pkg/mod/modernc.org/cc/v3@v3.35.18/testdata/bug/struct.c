struct outer {
	char magic;
	char pad1[3];
	union {
		char ceiling;
		char unused;
	};
	char pad2[3];
	char owner;
};

struct outer x = {
	.magic = 1,
	.pad1 = {2, 3, 4},
	.ceiling = 5,
	.pad2 = {6, 7, 8},
	.owner = 9,
};
