typedef unsigned char bool;
typedef unsigned char uint8_t;
typedef unsigned short uint16_t;
typedef unsigned uint32_t;

typedef struct {
	bool is_inline:1;
	bool visible:1;
	bool named:1;
	bool extra:1;
	bool has_changes:1;
	bool is_missing:1;
	bool is_keyword:1;
	uint8_t symbol;
	uint8_t padding_bytes;
	uint8_t size_bytes;
	uint8_t padding_columns;
	uint8_t padding_rows:4;
	uint8_t lookahead_bytes:4;
	uint16_t parse_state;
} SubtreeInlineData;

typedef struct {
	uint32_t child_count;
} SubtreeHeapData;

typedef union {
	SubtreeInlineData data;
	SubtreeHeapData *ptr;
} MutableSubtree;

int main() {
	MutableSubtree ms;
	ms = (MutableSubtree) { .data = {} };
	void *p = (void *)0;
	ms = (MutableSubtree) { .ptr = p };
	return ms.ptr != 0;
}
