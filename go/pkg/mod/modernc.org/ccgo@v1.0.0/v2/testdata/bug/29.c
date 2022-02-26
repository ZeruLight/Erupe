#include <assert.h>

int main() {
	int i = 42;
	for (int i = 10; i < 20; i++) {
		assert(i != 42);
	}
	assert(i == 42);

	i = 24;
	for (int i = 10; i < 20; i++) {
		assert(i != 42);
		assert(i != 24);
	}
	assert(i == 24);

	int j = 314;
	for (j = 0; j < 10; j++) {
		assert(j != 314);
	}
	assert(j == 10);

	int k = 278;
	for (j = 0; k < 300; k++) {
		int k = 1000;
		assert(k == 1000);
	}
	assert(k == 300);

	int l = 123;
	for (int l = 0; l < 10; l++) {
		int l = 1000;
		assert(l == 1000);
	}
	assert(l == 123);

	int m = 999;
	for (m = 0; m < 10; m++)
		assert(m != 999);

	assert(m == 10);

	int n = 888;
	for (int n = 0; n < 10; n++)
		assert(n != 888);

	assert(n == 888);
}
