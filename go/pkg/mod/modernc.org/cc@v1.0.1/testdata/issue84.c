#define a(b, ...) c(b, __VA_ARGS__)
a(1, 2, 3);
a(1, 2);
a(1);
