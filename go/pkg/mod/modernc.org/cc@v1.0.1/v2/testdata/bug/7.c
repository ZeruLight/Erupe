#include <assert.h>

int i;
int *pi = &i;
int **ppi = &pi;

int main() {
	assert(pi == &i);
	assert(ppi == &pi);
	assert((void*)pi != (void*)ppi);
	assert(&pi);
	assert(&ppi);

	int *spi = pi;
	int **sppi = ppi;

	assert(!i);
	assert(!*pi);
	assert(!**ppi);

	i = 42;
	assert(i == 42);
	assert(*pi == 42);
	assert(**ppi == 42);

	*pi = 24;
	assert(i == 24);
	assert(*pi == 24);
	assert(**ppi == 24);
	assert(pi == spi);
	assert(ppi == sppi);

	**ppi = 314;
	assert(i == 314);
	assert(*pi == 314);
	assert(**ppi == 314);
	assert(pi == spi);
	assert(ppi == sppi);
}
