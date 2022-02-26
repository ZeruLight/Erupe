#include <assert.h>
#include <stdarg.h>

va_list a;

void g(va_list ap) {
	assert(va_arg(ap, int) == 278);
}

void f(int i, ...) {
	va_list ap;
	va_start(ap, i);
	assert(va_arg(ap, int) == 42);
	assert(va_arg(ap, double) == 3.14);
	g(ap);
	assert(va_arg(ap, int) == 123);
	va_end(ap);
}

int main() {
	va_list b;
	__builtin_printf("%i %i\n", sizeof(a), sizeof(b));
	assert(sizeof(a) == sizeof(void*));
	assert(sizeof(b) == sizeof(void*));
	f(0, 42, 3.14, 278, 123);
}
