long __syscall(syscall_arg_t, ...);

#ifndef __scc
#define __scc(X) ((long) (X))
typedef long syscall_arg_t;
#endif


#define __SYSCALL_CONCAT(a,b) __SYSCALL_CONCAT_X(a,b)
#define __SYSCALL_CONCAT_X(a,b) a##b
#define __SYSCALL_DISP(b,...) __SYSCALL_CONCAT(b,__SYSCALL_NARGS(__VA_ARGS__))(__VA_ARGS__)
#define __SYSCALL_NARGS(...) __SYSCALL_NARGS_X(__VA_ARGS__,7,6,5,4,3,2,1,0,)
#define __SYSCALL_NARGS_X(a,b,c,d,e,f,g,h,n,...) n
#define __sys_open(...) __SYSCALL_DISP(__sys_open,,__VA_ARGS__)
#define __sys_open3(x,pn,fl,mo) __syscall4(SYS_openat, AT_FDCWD, pn, (fl)|O_LARGEFILE, mo)
#define __syscall4(n,a,b,c,d) (__syscall)(n,__scc(a),__scc(b),__scc(c),__scc(d))
#define sys_open(...) __syscall_ret(__sys_open(__VA_ARGS__))

int fd = sys_open(filename, flags, 0666);
