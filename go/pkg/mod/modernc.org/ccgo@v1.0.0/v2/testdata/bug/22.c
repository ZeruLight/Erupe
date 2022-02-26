#include <assert.h>

int main() {
	int i = 42;
	int j = 0;
	for (int i = 0; i < 3; i++) {
		j++;			
	}
	assert(i == 42);
	assert(j == 3);

	int k = 314;
	int *p = &k;
	for (int i = 0; i < 3; i++) {
		(*p)++;			
	}
	assert(k == 314+3);
}
