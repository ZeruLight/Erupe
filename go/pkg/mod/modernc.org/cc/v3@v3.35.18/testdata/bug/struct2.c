// https://gitlab.com/cznic/ccgo/-/issues/21#note_704291242

#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>

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
  uint32_t child_count;
} SubtreeHeapData;

typedef union {
  SubtreeInlineData data;
  const SubtreeHeapData *ptr;
} Subtree;

typedef struct {
  Subtree tree;
  uint32_t child_index;
  uint32_t byte_offset;
} StackEntry;

void myfunc() {
  Subtree tree;
  uint32_t next_index;
  uint32_t byte_offset;
  ((StackEntry) {
    .tree = ((tree).data.is_inline ? ((void *)0) : (Subtree *)((tree).ptr) - (tree).ptr->child_count)[next_index],
    .child_index = next_index,
    .byte_offset = byte_offset,
  });
  Subtree child = ((Subtree *)((tree).ptr) - (tree).ptr->child_count)[next_index];
  Subtree child2 = ((tree).data.is_inline ? ((void *)0) : (Subtree *)((tree).ptr) - (tree).ptr->child_count)[next_index];
  Subtree child3 = ((tree).data.is_inline ? (Subtree *)((tree).ptr) - (tree).ptr->child_count : ((void *)0))[next_index];
}
