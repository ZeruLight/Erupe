#include <assert.h>

static int i = 42;

static int j; // not produced

int f(int *p) { return *p; }

static void g() {} // not produced
static void h() {}

int main() {
	assert(f(&i) == 42);
	h();
}
