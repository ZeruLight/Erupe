static void f(int);

int main() {
	f(42);
	return 0;
}

static void f(int i) {
	int j = 0;
	if (j) {
		j++;
	}
}
