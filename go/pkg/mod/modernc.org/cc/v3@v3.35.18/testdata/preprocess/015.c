int f(int x) {
	char *p = "abc'def\"ghi";
	return x == 14 ? '"' : '\'';
}
