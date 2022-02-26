typedef struct TSLexer TSLexer;

struct TSLexer {
  int lookahead;
  int result_symbol;
  void (*advance)();
  void (*mark_end)();
  void (*get_column)();
  void (*is_at_included_range_start)();
  void (*eof)();
};

void f1() {}
void f2() {}
void f3() {}
void f4() {}
void f5() {}

int main() {
  TSLexer l = {
      .advance = f1,
      .mark_end = f2,
      .get_column = f3,
      .eof = f5,
      .is_at_included_range_start = f4,
      .lookahead = 0,
      .result_symbol = 0,
    };
  if (l.lookahead != 0) {
	  return __LINE__;
  }

  if (l.result_symbol != 0) {
	  return __LINE__;
  }

  if (l.advance != f1) {
	  return __LINE__;
  }

  if (l.mark_end != f2) {
	  return __LINE__;
  }

  if (l.get_column != f3) {
	  return __LINE__;
  }

  if (l.is_at_included_range_start != f4) {
	  return __LINE__;
  }

  if (l.eof != f5) {
	  return __LINE__;
  }

  return 0;
}
