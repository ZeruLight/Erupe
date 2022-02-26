#define TEST42 0x0502
#define TESTBASE 1
#define TEST3 (TEST42 >= 0x0502 || !defined (TESTBASE))
#define TEST4

#if TEST3 && defined (TEST4)
	int test = 4;
#endif


int main (void) {
	return test;
}
