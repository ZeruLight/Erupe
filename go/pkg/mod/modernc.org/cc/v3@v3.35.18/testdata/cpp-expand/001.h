#define f(x) f(2*x)
#define m(a) a(24)
f(f(42))
m(f)
