typedef int int_t;

// redeclaration error:
int foo(void *ptr);
typedef int foo(void *ptr);

// redeclaration error:
int_t foo2(void *ptr);
typedef int_t foo2(void *ptr);
