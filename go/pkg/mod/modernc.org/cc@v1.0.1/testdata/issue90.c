struct {
	unsigned a:1;
	unsigned b:1;
} x;

void f() {
	x.a + x.b;
}
