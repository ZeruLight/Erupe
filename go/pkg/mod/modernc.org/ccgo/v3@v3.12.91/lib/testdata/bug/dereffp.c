void f(int i) {}
void g(int i) {}
void x(int i) {}

int h(void (*proc)(int)) {
	if (proc != *proc) {
		__builtin_printf("oops\n");
		return 255;
	}

	if (*proc == f) {
		return 'f';
	}

	if (*proc == g) {
		return 'g';
	}

	return 0;
}

int h2(void (**proc)(int)) {
	if (*proc == f) {
		return 'f';
	}

	if (*proc == g) {
		return 'g';
	}

	return 0;
}

int (**gfpp)(int);
int (*gfp)(int);
int x2(int i) { return 2*i; }

int main() {
	if (h(f) != 'f') {
		return __LINE__;
	}

	if (h(&f) != 'f') {
		return __LINE__;
	}

	void (*p)(int);
	void (**q)(int) = &p;
	p = f;
	if (h2(q) != 'f') {
		return __LINE__;
	}

	p = &f;
	if (h2(q) != 'f') {
		return __LINE__;
	}

	if (h(g) != 'g') {
		return __LINE__;
	}

	if (h(&g) != 'g') {
		return __LINE__;
	}

	p = g;
	if (h2(q) != 'g') {
		return __LINE__;
	}

	p = &g;
	if (h2(q) != 'g') {
		return __LINE__;
	}

	// ----

	gfpp = &gfp;
	gfp = x2;
	if ((**gfpp)(__LINE__) != 2*__LINE__) {
		return __LINE__;
	}

	gfp = &x2;
	if ((**gfpp)(__LINE__) != 2*__LINE__) {
		return __LINE__;
	}

	int (*fp)(int);
	gfpp = &fp;
	fp = x2;
	if ((**gfpp)(__LINE__) != 2*__LINE__) {
		return __LINE__;
	}

	int (**fpp)(int) = &gfp;
	gfp = x2;
	if ((**fpp)(__LINE__) != 2*__LINE__) {
		return __LINE__;
	}

	gfp = &x2;
	if ((**fpp)(__LINE__) != 2*__LINE__) {
		return __LINE__;
	}

	fpp = &fp;
	fp = x2;
	if ((**fpp)(__LINE__) != 2*__LINE__) {
		return __LINE__;
	}
}
