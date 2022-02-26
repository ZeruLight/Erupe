int main() {
loop:
	for (int i = 0, j = 10; i < 3; i++) {
		__builtin_printf("%d %d\n", i, j);
		j++;
	}
	return 0;
	goto loop;
}
