#include <assert.h>

int f1() { return 1; }

int f2() { return 2; }

struct api {
	int(*a)();
	int(*b)();
} api, api2;

struct api *api3, *api4;

int main() {
	void *p = &api2;
	p = &api4;

	api.b = f1;
	assert(api.b() == 1);
	api2.b = f1;
	assert(api2.b() == 1);
	api3 = &api2;
	api3->b = f2;
	assert(api3->b() == 2);
	api4 = &api2;
	api4->b = f2;
	assert(api4->b() == 2);
}
