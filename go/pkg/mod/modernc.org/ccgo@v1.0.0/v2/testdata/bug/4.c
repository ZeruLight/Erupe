struct s {
	int i;
} x;

int foo(struct s *p) {
	return p->i;
}

int foo2(struct s *p) {
	return (*p).i;
}

int main() {
	x.i = 42;
	if (foo(&x) != 42) {
		abort();
	}
	if (foo2(&x) != 42) {
		abort();
	}
}
