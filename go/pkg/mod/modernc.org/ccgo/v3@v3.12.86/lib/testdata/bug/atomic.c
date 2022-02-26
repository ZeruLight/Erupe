_Atomic(int) i;
_Atomic int j;

int main() {
	i++;
	if (i != 1) {
		return __LINE__;
	}

	j++;
	if (j != 1) {
		return __LINE__;
	}

	return 0;
}
