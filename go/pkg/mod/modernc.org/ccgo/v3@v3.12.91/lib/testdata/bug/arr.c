struct a {
	int b;
	int c;
} a = {1, 2};

struct d {
	struct a e;
	int f;
};

struct d ga[1];

int main() {
	(*ga).e = a;
	if (ga->e.b != 1) {
		return __LINE__;
	}

	if (ga->e.c != 2) {
		return __LINE__;
	}

	struct d la[1] = {};
	(*la).e = a;

	if (la->e.b != 1) {
		return __LINE__;
	}

	if (la->e.c != 2) {
		return __LINE__;
	}

	return 0;
}
