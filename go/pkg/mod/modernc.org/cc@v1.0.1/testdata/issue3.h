#define FOO 3

void foo(int a, int b);

#define do_foo(x) \
  foo(x, FOO)
