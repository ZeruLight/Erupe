#define __signed signed

typedef __signed char __int8_t;

// for future tests
// (on the real code this producting another error)
// #if 5 > 1
// // ok?
// #endif

// remove this block -> works!
#if !defined(A) && (!defined(B) || defined(C))
// ok?
#endif

#define FOO (4)

// remove this block -> works!
#if (FOO) > 4
// ok?
#endif
