typedef struct {
	short a;
	unsigned short b:12;
	unsigned char c:1;
} s;

s f()
{
	return (s) {
		.b = 0xfef,
	};
}

s g()
{
	return (s) {
		.b = 0xfef,
		.c = 1,
	};
}

int main()
{
	s s;
	s = f();
	if (s.b != 0xfef) {
		return __LINE__;
	}

	if (s.c) {
		return __LINE__;
	}

	s = g();
	if (s.b != 0xfef) {
		return __LINE__;
	}

	if (!s.c) {
		return __LINE__;
	}

	return 0;
}
