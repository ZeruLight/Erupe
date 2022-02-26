#define foo(a, ...) bar(a, __VA_ARGS__)
foo(1, 2, 3);
