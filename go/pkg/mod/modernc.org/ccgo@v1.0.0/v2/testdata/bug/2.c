#include <stdlib.h>

void foo(char**) {}

int main() {
	char *p = NULL;
	foo(&p);
	return p != NULL;
}
