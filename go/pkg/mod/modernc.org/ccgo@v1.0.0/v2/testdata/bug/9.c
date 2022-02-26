#include <stdio.h>

int main() {
	float f32;
	unsigned long long u64;
	f32 = __FLT_MAX__;
	u64 = f32;
	printf("%f %llu\n", f32, u64);
	int i32;
	int *pi32;
	pi32 = &i32;
	*pi32 = 0x9.A70A99p+87;
	printf("%f %i\n", 0x9.A70A99p+87, i32);
	return 0;
}
