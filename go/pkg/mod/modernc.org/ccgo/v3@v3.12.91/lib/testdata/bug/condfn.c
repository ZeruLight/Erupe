int f(int n) {
	return 2*n;
}

int g(int n) {
	return 3*n;
}

int h(int c, int n) {
	return (c ? f : g)(n);
}

int i(int c, int n) {
	return (c ? &f : &g)(n);
}

int main() {
	if (h(0, 10) != 30) {
		return __LINE__;
	}

	if (h(1, 20) != 40) {
		return __LINE__;
	}
}
