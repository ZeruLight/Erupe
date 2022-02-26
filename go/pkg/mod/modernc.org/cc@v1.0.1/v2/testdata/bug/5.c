#include <assert.h>
#include <stdint.h>

#if defined(__PTRDIFF_TYPE__)
#define SQLITE_INT_TO_PTR1(X)  ((void*)(__PTRDIFF_TYPE__)(X))
#define SQLITE_PTR_TO_INT1(X)  ((int)(__PTRDIFF_TYPE__)(X))
#endif

#define SQLITE_INT_TO_PTR2(X)  ((void*)&((char*)0)[X])
#define SQLITE_PTR_TO_INT2(X)  ((int)(((char*)X)-(char*)0))

#define SQLITE_INT_TO_PTR3(X)  ((void*)(intptr_t)(X))
#define SQLITE_PTR_TO_INT3(X)  ((int)(intptr_t)(X))

#define SQLITE_INT_TO_PTR4(X)  ((void*)(X))
#define SQLITE_PTR_TO_INT4(X)  ((int)(X))

void foo(char *p, char *q) {
	int i, j = 42;
	char **z = &q;

	assert(p);
	assert(q);
	assert(z);
	p = 0;
	q = 0;
	z = 0;
	assert(!p);
	assert(!q);
	assert(!z);

#if defined(SQLITE_INT_TO_PTR1)
	p = 0;
	i = 0;
	p = SQLITE_INT_TO_PTR1(314);
	i = SQLITE_PTR_TO_INT1(p);
	assert(i == 314);
	p = SQLITE_INT_TO_PTR1(j);
	i = SQLITE_PTR_TO_INT1(p);
	assert(i == 42);
	q = 0;
	i = 0;
	q = SQLITE_INT_TO_PTR1(314);
	i = SQLITE_PTR_TO_INT1(q);
	assert(i == 314);
	q = SQLITE_INT_TO_PTR1(j);
	i = SQLITE_PTR_TO_INT1(q);
	assert(i == 42);
#endif

	p = 0;
	i = 0;
	p = SQLITE_INT_TO_PTR2(314);
	i = SQLITE_PTR_TO_INT2(p);
	assert(i == 314);
	p = SQLITE_INT_TO_PTR2(j);
	i = SQLITE_PTR_TO_INT2(p);
	assert(i == 42);
	q = 0;
	i = 0;
	q = SQLITE_INT_TO_PTR2(314);
	i = SQLITE_PTR_TO_INT2(q);
	assert(i == 314);
	q = SQLITE_INT_TO_PTR2(j);
	i = SQLITE_PTR_TO_INT2(q);
	assert(i == 42);

	p = 0;
	i = 0;
	p = SQLITE_INT_TO_PTR3(314);
	i = SQLITE_PTR_TO_INT3(p);
	assert(i == 314);
	p = SQLITE_INT_TO_PTR3(j);
	i = SQLITE_PTR_TO_INT3(p);
	assert(i == 42);
	q = 0;
	i = 0;
	q = SQLITE_INT_TO_PTR3(314);
	i = SQLITE_PTR_TO_INT3(q);
	assert(i == 314);
	q = SQLITE_INT_TO_PTR3(j);
	i = SQLITE_PTR_TO_INT3(q);
	assert(i == 42);

	p = 0;
	i = 0;
	p = SQLITE_INT_TO_PTR4(314);
	i = SQLITE_PTR_TO_INT4(p);
	assert(i == 314);
	p = SQLITE_INT_TO_PTR4(j);
	i = SQLITE_PTR_TO_INT4(p);
	assert(i == 42);
	q = 0;
	i = 0;
	q = SQLITE_INT_TO_PTR4(314);
	i = SQLITE_PTR_TO_INT4(q);
	assert(i == 314);
	q = SQLITE_INT_TO_PTR4(j);
	i = SQLITE_PTR_TO_INT4(q);
	assert(i == 42);
}

int main() {
	int i, j = 42;
	char c;
	char *p = &c, *q = &c;
	char **z = &q;

	foo(p, q);

	assert(p);
	assert(q);
	assert(z);
	p = 0;
	q = 0;
	z = 0;
	assert(!p);
	assert(!q);
	assert(!z);

#if defined(SQLITE_INT_TO_PTR1)
	p = 0;
	i = 0;
	p = SQLITE_INT_TO_PTR1(314);
	i = SQLITE_PTR_TO_INT1(p);
	assert(i == 314);
	p = SQLITE_INT_TO_PTR1(j);
	i = SQLITE_PTR_TO_INT1(p);
	assert(i == 42);
	q = 0;
	i = 0;
	q = SQLITE_INT_TO_PTR1(314);
	i = SQLITE_PTR_TO_INT1(q);
	assert(i == 314);
	q = SQLITE_INT_TO_PTR1(j);
	i = SQLITE_PTR_TO_INT1(q);
	assert(i == 42);
#endif

	p = 0;
	i = 0;
	p = SQLITE_INT_TO_PTR2(314);
	i = SQLITE_PTR_TO_INT2(p);
	assert(i == 314);
	p = SQLITE_INT_TO_PTR2(j);
	i = SQLITE_PTR_TO_INT2(p);
	assert(i == 42);
	q = 0;
	i = 0;
	q = SQLITE_INT_TO_PTR2(314);
	i = SQLITE_PTR_TO_INT2(q);
	assert(i == 314);
	q = SQLITE_INT_TO_PTR2(j);
	i = SQLITE_PTR_TO_INT2(q);
	assert(i == 42);

	p = 0;
	i = 0;
	p = SQLITE_INT_TO_PTR3(314);
	i = SQLITE_PTR_TO_INT3(p);
	assert(i == 314);
	p = SQLITE_INT_TO_PTR3(j);
	i = SQLITE_PTR_TO_INT3(p);
	assert(i == 42);
	q = 0;
	i = 0;
	q = SQLITE_INT_TO_PTR3(314);
	i = SQLITE_PTR_TO_INT3(q);
	assert(i == 314);
	q = SQLITE_INT_TO_PTR3(j);
	i = SQLITE_PTR_TO_INT3(q);
	assert(i == 42);

	p = 0;
	i = 0;
	p = SQLITE_INT_TO_PTR4(314);
	i = SQLITE_PTR_TO_INT4(p);
	assert(i == 314);
	p = SQLITE_INT_TO_PTR4(j);
	i = SQLITE_PTR_TO_INT4(p);
	assert(i == 42);
	q = 0;
	i = 0;
	q = SQLITE_INT_TO_PTR4(314);
	i = SQLITE_PTR_TO_INT4(q);
	assert(i == 314);
	q = SQLITE_INT_TO_PTR4(j);
	i = SQLITE_PTR_TO_INT4(q);
	assert(i == 42);
}
