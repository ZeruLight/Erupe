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
	.magic = '1',
	.pad1 = {'2', '3', '4'},
	.ceiling = '5',
	.pad2 = {'6', '7', '8'},
	.owner = '9',
};

char buf[10];

int main() {
	buf[0] = x.magic;
	buf[1] = x.pad1[0];
	buf[2] = x.pad1[1];
	buf[3] = x.pad1[2];
	buf[4] = x.ceiling;
	buf[5] = x.pad2[0];
	buf[6] = x.pad2[1];
	buf[7] = x.pad2[2];
	buf[8] = x.owner;
	__builtin_printf("%s\n", buf);
}
