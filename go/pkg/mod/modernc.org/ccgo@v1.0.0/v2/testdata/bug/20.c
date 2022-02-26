static int foo();

int (*var)() = foo;

int foo() { return 42; }

int main() {
	return var != foo || var() != 42;
}
