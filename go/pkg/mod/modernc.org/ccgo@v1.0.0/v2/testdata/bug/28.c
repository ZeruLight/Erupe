#include <stdlib.h>
#include <assert.h>

extern int a[];
int *b;
int c[5];

void f() {
	int c[5];

	for (int i = 0; i < 5; i++) {
		c[i] = 10*i+12;
	}

	for (int i = 0; i < 3; i++) {
		assert(a[i] == 10*i+10);
	}
	for (int i = 0; i < 4; i++) {
		assert(b[i] == 10*i+11);
	}
	for (int i = 0; i < 5; i++) {
		assert(c[i] == 10*i+12);
	}
}
 
int main() {
	b = malloc(4*sizeof (int));
	for (int i = 0; i < 4; i++) {
		b[i] = 10*i+11;
	}
	for (int i = 0; i < 5; i++) {
		c[i] = 10*i+12;
	}

	for (int i = 0; i < 3; i++) {
		assert(a[i] == 10*i+10);
	}
	for (int i = 0; i < 4; i++) {
		assert(b[i] == 10*i+11);
	}
	for (int i = 0; i < 5; i++) {
		assert(c[i] == 10*i+12);
	}
 	f();
}
 
int a[] = {10, 20, 30};
