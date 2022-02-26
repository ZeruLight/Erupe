typedef struct a *b;
#define c(type, name, paramName)                          \
  void set##name(const b node, type paramName); \
  type get##name(const b node);

c(void *, Context, context); // The final semicolon is a C++11 empty declaration, illegal in C99.

// See also http://en.cppreference.com/w/cpp/language/declarations
