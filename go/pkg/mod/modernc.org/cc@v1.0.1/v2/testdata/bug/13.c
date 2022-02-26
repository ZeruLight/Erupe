#include <assert.h>

typedef struct _OVERLAPPED {
	int Internal;
	int InternalHigh;
	union {
		struct {
			int Offset;
			int OffsetHigh;
		};
		void *Pointer;
	};
	unsigned hEvent;
} OVERLAPPED;

OVERLAPPED test;
OVERLAPPED test2;

int main() {
	OVERLAPPED test3, test4;
	OVERLAPPED *p = &test2, *q = &test4;
	test.Offset = 42;
	p->Offset = 43;
	test3.Offset = 44;
	q->Offset = 45;
	assert(test.Offset == 42);
	assert(p->Offset == 43);
	assert(test3.Offset == 44);
	assert(q->Offset == 45);
}
