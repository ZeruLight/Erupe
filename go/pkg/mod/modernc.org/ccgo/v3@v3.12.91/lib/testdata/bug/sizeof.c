#if (__SIZEOF_INT__ == 4)
typedef int int32;
#elif (__SIZEOF_LONG__ == 4)
typedef long int32;
#else
#error Add target support for int32
#endif

union u {
  char c[5];
  int32 i;
} u;

int main() {
	__builtin_printf("%i\n", sizeof(u));
}
