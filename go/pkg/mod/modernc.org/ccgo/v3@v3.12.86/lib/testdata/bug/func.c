typedef int (f) (int i);

int g(f func) {
	return func(42);
}

int h(int i) {
	return i+1;
}

int main() {
	__builtin_printf("%i\n", g(h));
}
