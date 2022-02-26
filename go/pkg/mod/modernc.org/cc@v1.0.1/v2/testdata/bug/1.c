#include <assert.h>
extern void abort (void);

void f1(int *ip) {
	*ip = 24;
}

void t1() {
  	int i = 42;
	assert(i == 42);
	f1(&i);
	assert(i == 24);
}

void t2() {
	int i = 42;
	int *ip = &i;
	assert(i == 42);
	assert(*ip == 42);
	assert(ip == &i);
	f1(ip);
	assert(i == 24);
}

void f3(int **ipp) {
	**ipp = 24;
}

void t3() {
	int i = 42;
	int *ip = &i;
	assert(i == 42);
	assert(*ip == 42);
	assert(ip == &i);
	f3(&ip);
	assert(i == 24);
}

void f4(int **ipp) {
}

void t4() {
	int i = 42;
	int *ip = &i;
	assert(i == 42);
	assert(*ip == 42);
	assert(ip == &i);
	f4(&ip);
}

int main (void) {
	t1();
	t2();
	t3();
	t4();
	return 0;
}
