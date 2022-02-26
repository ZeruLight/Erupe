#define QUOTE_(s) #s
#define QUOTE(s) QUOTE_(s)
#define check(t) check(QUOTE(t), __alignof__(t))
 
check (void);
