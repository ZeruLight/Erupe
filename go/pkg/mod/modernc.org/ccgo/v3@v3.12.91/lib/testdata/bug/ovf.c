#define FOO 0xffffffff

int x = FOO;

int main() {
	__builtin_printf("%i\n", x);
}
