#include <stdint.h>

typedef union {
	struct {
		char type;
		short state;
		char extra;
		char repetition;
	} shift;
	struct {
		char type;
		char child_count;
		short symbol;
		short dynamic_precedence;
		short production_id;
	} reduce;
	char type;
} TSParseAction;

typedef union {
	TSParseAction action;
	struct {
		char count;
		char reusable;
	} entry;
} TSParseActionEntry;

static const TSParseActionEntry ts_parse_actions[] = {
	[13] = {
			.entry = {
				.count = 11,
				.reusable = 22
			}
		},
		{
			{
				.reduce = {
					.type = 33,
					.symbol = 44,
					.child_count = 55,
				},
			}
		},
};

static const TSParseActionEntry ts_parse_actions2[] = {
	[13] = {
			{
				.reduce = {
					.type = 66,
					.symbol = 77,
					.child_count = 88,
				},
			}
		},
};

int main() {
	if (ts_parse_actions[13].entry.count != 11) {
		return __LINE__;
	}

	if (ts_parse_actions[13].entry.reusable != 22) {
		return __LINE__;
	}

	if (ts_parse_actions[14].action.reduce.type != 33) {
		return __LINE__;
	}

	if (ts_parse_actions[14].action.reduce.symbol != 44) {
		return __LINE__;
	}

	if (ts_parse_actions[14].action.reduce.child_count != 55) {
		return __LINE__;
	}

	if (ts_parse_actions2[13].action.reduce.type != 66) {
		return __LINE__;
	}

	if (ts_parse_actions2[13].action.reduce.symbol != 77) {
		return __LINE__;
	}

	if (ts_parse_actions2[13].action.reduce.child_count != 88) {
		return __LINE__;
	}

	return 0;
}
