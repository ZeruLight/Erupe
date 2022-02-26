typedef long long FT_Fixed;

#define af_floatToFixed( f ) \
          ( (FT_Fixed)( (f) * 65536.0 + 0.5 ) )

int main() {
    long long x = af_floatToFixed( .01 );
    __builtin_printf("%llu\n", x);
    if (x != 655) {
	    return __LINE__;
    }

    return 0;
}
