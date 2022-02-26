// #include <stdint.h>

// The definition used by cxgo

#define uint8_t _cxgo_uint8
#define _cxgo_uint8  unsigned __int8

int main() {
	typedef uint8_t MYubyte;
	MYubyte vendor[] = "Robert Winkler";
	return 0;
}
