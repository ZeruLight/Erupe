#define m(x) #x
#define n(x, y) #x
#define o(x, y) #y
#define p(x, y, z) #y

char s[] = m();
char t[] = n(, 42);
char u[] = o(42,);
char v[] = p(42,,314);
