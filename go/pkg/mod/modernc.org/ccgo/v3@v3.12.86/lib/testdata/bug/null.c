int f(char * p) {
	return p != 0;
}

char c;

int main() {
	if (!f(&c)) {
		return __LINE__;
	}

	if (f('\0')) {
		return __LINE__;
	}

	return 0;
}
