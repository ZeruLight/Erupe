#include <assert.h>

void foo(int *v[]) {
	assert(*v[0] == 42 || *v[0] == 314);
}

void bar(int v[]) {
	assert(v[0] == 42 || v[0] == 314);
}

int a[] = {42};

int main() {
	int *p = a;
	foo(&p);
	bar(a);
	int b[] = {314};
	p = b;
	foo(&p);
	bar(b);
}
