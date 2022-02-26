void f(int p[], int q[]) {
	int i;
	for (i = 0; i < 4; i++) {
		__builtin_printf("%i: %i\n", i, p[i]);
	}
	int **pp = &p;
	*pp = q;
	for (i = 0; i < 4; i++) {
		__builtin_printf("%i: %i\n", i, q[i]);
	}
	for (i = 0; i < 4; i++) {
		__builtin_printf("%i: %i\n", i, p[i]);
	}
}

int p[] = {1, 2, 3, 4};
int q[] = {10, 20, 30, 40};

int main() {
	f(p, q);
}
