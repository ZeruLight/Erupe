#include <stdlib.h>
#include <stdio.h>

typedef char *(foo)(int, void*);

static foo bar;

int main() {
       foo * f = bar;
       printf("%s\n", bar(42, 0));
       printf("%s\n", f(42, 0));
}

static char *bar(int i, void *p) {
       if (i != 42 || p) {
               abort();
       }
       return "ok";
}
