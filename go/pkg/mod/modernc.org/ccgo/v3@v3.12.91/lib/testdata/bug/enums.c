enum {
	a=100, b, c
} x;

struct {
	int i;
	enum {d = 200, e} en;
} y;

int main() {
	__builtin_printf("%i\n", a);
	__builtin_printf("%i\n", b);
	__builtin_printf("%i\n", c);
	__builtin_printf("%i\n", d);
	__builtin_printf("%i\n", e);
	enum {a = 300, b, c} x;
	enum {d = 400, e} en;
	__builtin_printf("%i\n", a);
	__builtin_printf("%i\n", b);
	__builtin_printf("%i\n", c);
	__builtin_printf("%i\n", d);
	__builtin_printf("%i\n", e);
}
