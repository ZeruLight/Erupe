// https://gitlab.com/cznic/sqlite2go/issues/9

typedef void *HWND;

typedef struct _RPC_ASYNC_STATE {
	union {
		struct {
			HWND hWnd;
			int Msg;
		} HWND; // Both a field name and a typedef name.
	} u;
} RPC_ASYNC_STATE;

int main() {}
