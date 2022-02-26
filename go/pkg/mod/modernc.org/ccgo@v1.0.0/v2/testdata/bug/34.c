int main() {
	double x = 1.0;
	double inf = x / 0.0; // +Inf
	double nan = inf/inf; // NaN
	if (nan < 1e-8) {
		return 1;
	}
}
