#include <assert.h>
#include <errno.h>

int foo() { return 42; }

int main() {
	int j = foo();
	__errno_location();
	int *p = __errno_location();
	int i = *__errno_location();
	p = &*__errno_location();
	errno = 42;
	assert(*p == 42);
	p = &(*__errno_location());
	errno = 421;
	assert(*p == 421);
}
