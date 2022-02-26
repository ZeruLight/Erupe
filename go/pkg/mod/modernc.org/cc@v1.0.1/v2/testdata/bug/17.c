#include <assert.h>

int main() {
	int i = 42;
	switch (i) {
		int j = 314;
	case 41:
		assert(0);
		break;
	case 42:
		assert(i == 42);
		assert(j == 0); // Not guaranteed in C, only in the Go translation.
		break;
	case 43:
		assert(0);
		break;
	default:
		assert(0);
		break;
	}
	assert(i == 42);
}
