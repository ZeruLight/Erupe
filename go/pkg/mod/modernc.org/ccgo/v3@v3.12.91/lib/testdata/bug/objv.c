struct obj {
	void *a;
	void *b;
};

void g(void *p) {}

struct obj **f(struct obj **objv) {
	g(*objv++);
	return objv;
}

struct obj **f2(struct obj *objv[]) {
	// Bug: increment objv by sizeof obj.
	// Should increment by sizeof obj*.
	g(*objv++);
	return objv;
}

int main() {
	struct obj obj;
	struct obj* objv[] = {&obj, &obj};
	__builtin_printf("%d %d\n", f(objv) == &objv[1], f2(objv) == &objv[1]);
	__builtin_printf("%d %d\n", f(objv) == objv+1, f2(objv) == objv+1);
}
