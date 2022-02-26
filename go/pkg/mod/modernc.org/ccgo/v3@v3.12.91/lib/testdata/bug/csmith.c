struct S2 {
   short  f0;
   unsigned f1: 3;
   signed f2 : 24;
} x;

int main() {
	x.f1 = 7;
	x.f2 = 0x2aaaaa;
	__builtin_printf("%i %i\n", x.f1, x.f2);
	int i = x.f2 |= 1;
	__builtin_printf("%i %i\n", x.f1, x.f2);
	__builtin_printf("%i\n", i);
	x.f1 = 0;
	__builtin_printf("%i %i\n", x.f1, x.f2);
	x.f2 = 0x555555;
	__builtin_printf("%i %i\n", x.f1, x.f2);
	i = x.f2 |= 2;
	__builtin_printf("%i %i\n", x.f1, x.f2);
	__builtin_printf("%i\n", i);
}

