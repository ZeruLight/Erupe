int f(int i) {
	return 2*i;
}

int g(int i) {
	return 10*i;
}

typedef int (*intf)(int);

intf *ga[2] = {f, g};
intf *gap = &ga[0];

int main() {
	intf la[2] = {f, g};
	intf *lap = &la[0];

	intf x = *gap;
	if (x(1) != 2) {
		return 1;
	}

	x = ga[0];
	if (x(2) != 4) {
		return 2;
	}

	x = ga[1];
	if (x(3) != 30) {
		return 3;
	}

	x = *lap;
	if (x(3) != 6) {
		return 4;
	}

	x = la[0];
	if (x(4) != 8) {
		return 5;
	}

	x = la[1];
	if (x(5) != 50) {
		return 6;
	}

	x = *gap++;
	if (x(6) != 12) {
		return 7;
	}

	x = *gap++;
	if (x(7) != 70) {
		return 8;
	}

	x = *lap++;
	if (x(8) != 16) {
		return 9;
	}

	x = *lap++;
	if (x(9) != 90) {
		return 10;
	}

	gap = &ga[0];
	if ((*gap++)(10) != 20) {
		return 11;
	}

	if ((*gap++)(11) != 110) {
		return 12;
	}

	lap = &la[0];
	if ((*lap++)(10) != 20) {
		return 13;
	}

	if ((*lap++)(40) != 400) {
		return 14;
	}

	return 0;
}
