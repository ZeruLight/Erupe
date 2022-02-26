int main() {
	int i = 1, j = 2, k;
	k = i, j;
	if (k != 1)
		abort();
	k = (i, j);
	if (k != 2)
		abort();
}
