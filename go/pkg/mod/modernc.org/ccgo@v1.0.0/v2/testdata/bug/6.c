#include <assert.h>

char *tld_p;
char *esc_tld_p;

int main() {
	assert(!tld_p);
	assert(!esc_tld_p);
	char **pp = &esc_tld_p;
	assert(pp);

	static char *static_p;
	static char *esc_static_p;

	assert(!static_p);
	assert(!esc_static_p);
	pp = &esc_static_p;
	assert(pp);
}
