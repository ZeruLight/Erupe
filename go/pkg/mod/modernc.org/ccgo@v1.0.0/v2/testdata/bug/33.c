int main() {
	int zero = 0, one = 1;
	if (0)
		printf("4\n");
	if (1)
		printf("6\n");
	if (zero)
		printf("8\n");
	if (one)
		printf("10\n");

	if (0) {
		printf("13\n");
	}
	if (1) {
		printf("16\n");
	}
	if (zero) {
		printf("19\n");
	}
	if (one) {
		printf("21\n");
	}

	if (0)
		printf("26\n");
	else
		printf("28\n");
	if (1)
		printf("30\n");
	else
		printf("32\n");
	if (zero)
		printf("34\n");
	else
		printf("36\n");
	if (one)
		printf("38\n");
	else
		printf("40\n");

	if (0) {
		printf("43\n");
	} else {
		printf("45\n");
	}
	if (1) {
		printf("48\n");
	} else {
		printf("50\n");
	}
	if (zero) {
		printf("52\n");
	} else {
		printf("55\n");
	}
	if (one) {
		printf("58\n");
	} else {
		printf("60\n");
	}

	int e = 109;
	if (e==109) e=-1;
	else if (e==110) e=109;
	printf("e %i\n", e);
	e = 110;
	if (e==109) e=-1;
	else if (e==110) e=109;
	printf("e %i\n", e);
}
