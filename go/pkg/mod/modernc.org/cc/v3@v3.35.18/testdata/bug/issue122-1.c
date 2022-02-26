#include <stdio.h>
#include <stdlib.h>

struct inner {
   int f;
   int g;
   int h;
};

struct outer {
   int A;
   struct inner B;
   int C;
};

struct outer x = {
   .C = 55,
   .B.g = 33,
   .A = 11,
   .B.h = 44,
   .B.f = 22,
};

struct outer y = {
   .C = 55,
   .B.g = 33, 44,
   .A = 11,
   .B.f = 22,
};

int main() {
	if (x.A != 11) {
		printf("x.A %i\n", x.A);
		abort();
	}
	if (x.B.f != 22) {
		printf("x.B.f %i\n", x.B.f);
		abort();
	}
	if (x.B.g != 33) {
		printf("x.B.g %i\n", x.B.g);
		abort();
	}
	if (x.B.h != 44) {
		printf("x.B.f %i\n", x.B.h);
		abort();
	}
	if (x.C != 55) {
		printf("x.C %i\n", x.C);
		abort();
	}
	if (y.A != 11) {
		printf("y.A %i\n", x.A);
		abort();
	}
	if (y.B.f != 22) {
		printf("y.B.f %i\n", x.B.f);
		abort();
	}
	if (y.B.g != 33) {
		printf("y.B.g %i\n", x.B.g);
		abort();
	}
	if (y.B.h != 44) {
		printf("y.B.f %i\n", x.B.h);
		abort();
	}
	if (y.C != 55) {
		printf("y.C %i\n", x.C);
		abort();
	}
}
