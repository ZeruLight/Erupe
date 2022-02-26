#define assert(x) ((void)((x) ? 0 : (__builtin_abort(), 0)))

int main() {
	int i = 1;
	assert(i);
}
