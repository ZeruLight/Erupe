#define a(b, c, ...) d(c)
a(1, 2, 3)
#undef a
#define a(b, c...) d(c)
a(1, 2, 3)
