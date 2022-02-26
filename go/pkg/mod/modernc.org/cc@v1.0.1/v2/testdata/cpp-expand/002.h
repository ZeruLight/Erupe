#define x      3
#define f(a)   f(x * (a))
#undef  x      
#define x      2
#define g      f
#define w      0,1

g(x+(3,4)-w)
