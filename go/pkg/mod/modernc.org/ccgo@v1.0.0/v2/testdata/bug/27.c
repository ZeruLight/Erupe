#include <assert.h>

int main() {
	int a[9] = { 0, 1, 2, 3, 4, 5, 6, 7, 8 };
	assert(sizeof (a) == 36);
	assert(sizeof (*a) == 4);
  	assert(sizeof (a) / sizeof (*a) == 9);
}
