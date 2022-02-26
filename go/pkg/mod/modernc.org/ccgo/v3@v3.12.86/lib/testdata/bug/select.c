union s {
	int a;
	int b;
} s = {42};

union s f() {
	return s;
}

int main() {
	if (f().a != 42) {
		return __LINE__;
	}

	if (f().b != 42) {
		return __LINE__;
	}

	return 0;
}
