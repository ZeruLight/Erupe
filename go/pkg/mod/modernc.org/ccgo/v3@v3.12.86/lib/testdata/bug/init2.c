#include <stdbool.h>
#include <stdint.h>

typedef struct TSLexer TSLexer;
typedef int Length;
typedef struct{ int x; } TSRange;
typedef struct{ int y; } TSInput;
typedef struct{ int z; } TSLogger;

struct TSLexer {
  int lookahead;
  int result_symbol;
  void (*advance)(TSLexer *, bool);
  void (*mark_end)(TSLexer *);
  uint32_t (*get_column)(TSLexer *);
  bool (*is_at_included_range_start)(const TSLexer *);
  bool (*eof)(const TSLexer *);
};

typedef struct {
  TSLexer data;
  Length current_position;
  Length token_start_position;
  Length token_end_position;

  TSRange *included_ranges;
  const char *chunk;
  TSInput input;
  TSLogger logger;

  uint32_t included_range_count;
  uint32_t current_included_range_index;
  uint32_t chunk_start;
  uint32_t chunk_size;
  uint32_t lookahead_size;
  bool did_get_column;

  char debug_buffer[100];
} Lexer;

void advance(TSLexer *a, bool b) {}
void mark_end(TSLexer *a) {}
uint32_t get_column(TSLexer * a) { return 0; };
bool is_at_included_range_start(const TSLexer *a) { return 0; }
bool eof(const TSLexer *a) { return 0; }

void ts_lexer_init(Lexer *self) {
  *self = (Lexer) {
    .data = {
      // The lexer's methods are stored as struct fields so that generated
      // parsers can call them without needing to be linked against this
      // library.
      .advance = advance,
      .mark_end = mark_end,
      .get_column = get_column,
      .is_at_included_range_start = is_at_included_range_start,
      .eof = eof,
      .lookahead = 42,
      .result_symbol = 43,
    },
    .current_position = 44,
  };
}

int main() {
  Lexer l;
  ts_lexer_init(&l);
  if (l.data.lookahead != 42) {
	  return __LINE__;
  }

  if (l.data.result_symbol != 43) {
	  return __LINE__;
  }

  if (l.data.advance != advance) {
	  return __LINE__;
  }

  if (l.data.mark_end != mark_end) {
	  return __LINE__;
  }

  if (l.data.get_column != get_column) {
	  return __LINE__;
  }

  if (l.data.is_at_included_range_start != is_at_included_range_start) {
	  return __LINE__;
  }

  if (l.data.eof != eof) {
	  return __LINE__;
  }

  return 0;
}
