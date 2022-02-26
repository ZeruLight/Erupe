#include <stdbool.h>
#include <stdint.h>

typedef int Length, TSSymbol, TSStateId;

typedef struct {
  bool is_inline : 1;
  bool visible : 1;
  bool named : 1;
  bool extra : 1;
  bool has_changes : 1;
  bool is_missing : 1;
  bool is_keyword : 1;
  uint8_t symbol;
  uint8_t padding_bytes;
  uint8_t size_bytes;
  uint8_t padding_columns;
  uint8_t padding_rows : 4;
  uint8_t lookahead_bytes : 4;
  uint16_t parse_state;
} SubtreeInlineData;

typedef struct {
  union {
    char *long_data;
    char short_data[24];
  };
  uint32_t length;
} ExternalScannerState;

typedef struct {
  volatile uint32_t ref_count;
  Length padding;
  Length size;
  uint32_t lookahead_bytes;
  uint32_t error_cost;
  uint32_t child_count;
  TSSymbol symbol;
  TSStateId parse_state;

  bool visible : 1;
  bool named : 1;
  bool extra : 1;
  bool fragile_left : 1;
  bool fragile_right : 1;
  bool has_changes : 1;
  bool has_external_tokens : 1;
  bool depends_on_column: 1;
  bool is_missing : 1;
  bool is_keyword : 1;

  union {
    // Non-terminal subtrees (`child_count > 0`)
    struct {
      uint32_t visible_child_count;
      uint32_t named_child_count;
      uint32_t node_count;
      uint32_t repeat_depth;
      int32_t dynamic_precedence;
      uint16_t production_id;
      struct {
        TSSymbol symbol;
        TSStateId parse_state;
      } first_leaf;
    };

    // External terminal subtrees (`child_count == 0 && has_external_tokens`)
    ExternalScannerState external_scanner_state;

    // Error terminal subtrees (`child_count == 0 && symbol == ts_builtin_sym_error`)
    int32_t lookahead_char;
  };
} SubtreeHeapData;

typedef union {
  SubtreeInlineData data;
  const SubtreeHeapData *ptr;
} Subtree;

int main() {
	Subtree external_token;
	SubtreeHeapData hd;
	external_token.ptr = &hd;
	ExternalScannerState ss;
	ss.length = 42;
	hd.external_scanner_state = ss;
	int i;
	i = external_token.ptr
		->external_scanner_state.length;
	if (i != 42) {
		return __LINE__;
	}

	return 0;
}
