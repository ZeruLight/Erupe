union u {
	char *a;
	int b;
};

union u f(int i) {
	return (union u){b: i};
}

union u g(int i) {
	return (union u){.b = i};
}

int main() {
	int i = 0;
	union u u, v;
	while ((u = f(i)).b < 5) {
		__builtin_printf("%d\n", i++);
	}
	__builtin_printf("%d\n", u.b);
	i = 0;
	while ((v = g(i)).b < 3) {
		__builtin_printf("%d\n", i++);
	}
	__builtin_printf("%d\n", v.b);
	return 0;
}
