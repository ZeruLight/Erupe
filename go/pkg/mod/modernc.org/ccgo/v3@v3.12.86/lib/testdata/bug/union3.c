typedef int Length;
typedef int TSStateId;
typedef int int32_t;
typedef unsigned char bool;
typedef unsigned char uint8_t;
typedef unsigned short uint16_t;
typedef unsigned uint32_t;
typedef void* TSSymbol;

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
        void* symbol;
        int parse_state;
      } first_leaf;
    };

    // External terminal subtrees (`child_count == 0 && has_external_tokens`)
    int external_scanner_state;

    // Error terminal subtrees (`child_count == 0 && symbol == ts_builtin_sym_error`)
    int32_t lookahead_char;
  };
} SubtreeHeapData;

typedef union {
  SubtreeInlineData data;
  const SubtreeHeapData *ptr;
} Subtree;

static inline uint32_t ts_subtree_repeat_depth(Subtree self) {
  return self.data.is_inline ? 0 : self.ptr->repeat_depth;
}

int main() {
	SubtreeHeapData data;
	data.repeat_depth = 42;
	Subtree tree;
	tree.ptr = &data;
	return ts_subtree_repeat_depth(tree) != 42;
}
