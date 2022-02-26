#include <assert.h>

typedef struct _SYSTEM_INFO {
	union {
		int dwOemId;
		struct {
			int wProcessorArchitecture;
			int wReserved;
		};
	};
	int dwPageSize;
	int lpMinimumApplicationAddress;
	int lpMaximumApplicationAddress;
	int dwActiveProcessorMask;
	int dwNumberOfProcessors;
	int dwProcessorType;
	int dwAllocationGranularity;
	int wProcessorLevel;
	int wProcessorRevision;
} SYSTEM_INFO;

int main() {
	SYSTEM_INFO sysinfo;
	sysinfo.dwPageSize = 1;
}
