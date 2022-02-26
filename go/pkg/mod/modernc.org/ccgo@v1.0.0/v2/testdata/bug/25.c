#include <stddef.h>
#include <assert.h>

int i;
int ei;

char *pc;

char arr[42];

int main() {
	assert(i == 0);

	int *ip = &ei;
	assert(ei == 0);

	char *charp;
	charp = pc;
	assert(pc == NULL);
	assert(!pc);

	charp = arr;

	int m3;
	m3 = -3;
	assert(m3 == -3);
	assert(-3 == m3);

	unsigned u3;
	u3 = -3;
	assert(u3 == -3);
	assert(-3 == u3);
}
