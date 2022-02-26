#define NO_NUMBER (((long long) (~ (unsigned) 0)) + 1)

int main() {
	__builtin_printf("a) %x\n", (unsigned)0);
	__builtin_printf("b) %x\n", (~(unsigned)0));
	__builtin_printf("c) %lli\n", (long long)(~(unsigned)0));
	__builtin_printf("d) %lli\n", ((long long)(~(unsigned)0))+1);
	__builtin_printf("e) %lli\n", (long long)NO_NUMBER);
	if (((int) NO_NUMBER) != 0 || NO_NUMBER == 0) {
		__builtin_abort();
	}
}

