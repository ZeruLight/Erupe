// +build ignore

#include <features.h>
#include <stdio.h>

int main()
{
#ifdef _POSIX_SOURCE
	printf("_POSIX_SOURCE %li\n", (long)_POSIX_SOURCE);
#endif
#ifdef _POSIX_C_SOURCE
	printf("_POSIX_C_SOURCE %li\n", (long)_POSIX_C_SOURCE);
#endif
#ifdef _XOPEN_SOURCE
	printf("_XOPEN_SOURCE %li\n", (long)_XOPEN_SOURCE);
#endif
#ifdef _XOPEN_SOURCE_EXTENDED
	printf("_XOPEN_SOURCE_EXTENDED %li\n", (long)_XOPEN_SOURCE_EXTENDED);
#endif
#ifdef _LARGEFILE_SOURCE
	printf("_LARGEFILE_SOURCE %li\n", (long)_LARGEFILE_SOURCE);
#endif
#ifdef _LARGEFILE64_SOURCE
	printf("_LARGEFILE64_SOURCE %li\n", (long)_LARGEFILE64_SOURCE);
#endif
#ifdef _FILE_OFFSET_BITS
	printf("_FILE_OFFSET_BITS %li\n", (long)_FILE_OFFSET_BITS);
#endif
#ifdef _ISOC99_SOURCE
	printf("_ISOC99_SOURCE %li\n", (long)_ISOC99_SOURCE);
#endif
#ifdef __STDC_WANT_LIB_EXT2__
	printf("__STDC_WANT_LIB_EXT2__ %li\n", (long)__STDC_WANT_LIB_EXT2__);
#endif
#ifdef __STDC_WANT_IEC_60559_BFP_EXT__
	printf("__STDC_WANT_IEC_60559_BFP_EXT__ %li\n", (long)__STDC_WANT_IEC_60559_BFP_EXT__);
#endif
#ifdef __STDC_WANT_IEC_60559_FUNCS_EXT__
	printf("__STDC_WANT_IEC_60559_FUNCS_EXT__ %li\n", (long)__STDC_WANT_IEC_60559_FUNCS_EXT__);
#endif
#ifdef __STDC_WANT_IEC_60559_TYPES_EXT__
	printf("__STDC_WANT_IEC_60559_TYPES_EXT__ %li\n", (long)__STDC_WANT_IEC_60559_TYPES_EXT__);
#endif
#ifdef _GNU_SOURCE
	printf("_GNU_SOURCE %li\n", (long)_GNU_SOURCE);
#endif
#ifdef _DEFAULT_SOURCE
	printf("_DEFAULT_SOURCE %li\n", (long)_DEFAULT_SOURCE);
#endif
#ifdef _REENTRANT
	printf("_REENTRANT %li\n", (long)_REENTRANT);
#endif
#ifdef _THREAD_SAFE
	printf("_THREAD_SAFE %li\n", (long)_THREAD_SAFE);
#endif
}
