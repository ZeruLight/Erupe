int f(int n) {
	return 2*n;
}

int (*fp1)(int) = f;
int (*fp2)(int) = &f;

int main() {
	__builtin_printf("%i\n", fp1(10));
	__builtin_printf("%i\n", (*fp1)(20));
	__builtin_printf("%i\n", (**fp1)(30));
	__builtin_printf("%i\n", fp2(40));
	__builtin_printf("%i\n", (*fp2)(50));
	__builtin_printf("%i\n", (**fp2)(60));

	int (*p1)(int) = f;
	int (*p2)(int) = &f;

	__builtin_printf("%i\n", p1(11));
	__builtin_printf("%i\n", (*p1)(21));
	__builtin_printf("%i\n", (**p1)(31));
	__builtin_printf("%i\n", p2(41));
	__builtin_printf("%i\n", (*p2)(51));
	__builtin_printf("%i\n", (**p2)(61));

	int (*q1)(int) = f;
	int (*q2)(int) = &f;
	void *p = &q1;
	void *q = &q2;

	__builtin_printf("%i\n", q1(12));
	__builtin_printf("%i\n", (*q1)(22));
	__builtin_printf("%i\n", (**q1)(32));
	__builtin_printf("%i\n", q2(42));
	__builtin_printf("%i\n", (*q2)(52));
	__builtin_printf("%i\n", (**q2)(62));
}
