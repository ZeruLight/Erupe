#include <assert.h>

void f(int g, int e) {
	assert(g == e);
}

int main() {
	int match = 0;
	f(match ? 0 : 1, 1);
	match = 1;
	f(match ? 0 : 1, 0);
}
