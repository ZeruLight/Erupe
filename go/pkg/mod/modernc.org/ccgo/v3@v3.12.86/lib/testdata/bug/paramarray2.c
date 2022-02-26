void f(int p[]) {
	int *q = &p[2];
	__builtin_printf("%i\n", *q);
}

struct s {
	int x, y;
};

void g(struct s p[]) {
	int *q = &p[1].y;
	__builtin_printf("%i\n", *q);
}

void h(struct s p[]) {
	int *q = &p->y;
	__builtin_printf("%i\n", *q);
}

int main() {
	int a[] = {1, 2, 42, 3, 4};
	f(a);
	struct s b[] = {{1, 2}, {3, 4}, {5, 6}};
	g(b);
	h(b);
}
