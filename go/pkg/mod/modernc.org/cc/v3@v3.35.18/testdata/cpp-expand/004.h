#define __str(x) # x
#define ____header(name, os, arch) __str(name ## _ ## os ## _ ## arch.h)
#define ___header(name, os, arch) ____header(name, os, arch)
#define __header(name) ___header(name, __os__, __arch__)

#define bug1(name) __str(name ## _ ## os)
bug1(a);

#define __os__ linux
#define __arch__ amd64

____header(a, b, c);
___header(a, b, c);
__header(a);

#define __str2(x) #x
#define ____header2(name, os, arch) __str2(name##_##os##_##arch.h)
#define ___header2(name, os, arch) ____header2(name, os, arch)
#define __header2(name) ___header2(name, __os__, __arch__)

#define bug2(name) __str2(name##_##os)
bug2(a);

____header2(a, b, c);
___header2(a, b, c);
__header2(a);
