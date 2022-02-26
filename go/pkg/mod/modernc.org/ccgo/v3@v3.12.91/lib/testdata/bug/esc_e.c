#include <stdio.h>

int main() {
	char *s = "abc\edef";
	int c;
	while (c = *s++) {
		printf("%d\n", c);
	}
}
