#include <assert.h>

#define __mingw_choose_expr __builtin_choose_expr

int main() {
	int i = __mingw_choose_expr(0, 2, 3);
	int j = __mingw_choose_expr(1, 2, 3);
	assert(i == 3);
	assert(j == 2);
}
