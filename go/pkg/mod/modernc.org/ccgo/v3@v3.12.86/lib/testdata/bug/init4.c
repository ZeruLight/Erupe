typedef union {
	struct {
		char type;
		short state;
		char extra;
		char repetition;
	} shift;
	struct {
		char child_count;
		char type;
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
	[14] =
		{
			{
				.reduce = {
					.type = 33,
					.dynamic_precedence = 44,
				},
			}
		},
};

int main() {
	if (sizeof(TSParseAction) != 8) {
		return __LINE__;
	}

	if (sizeof ts_parse_actions / sizeof(TSParseAction) != 15) {
		return __LINE__;
	}

	if (ts_parse_actions[14].action.reduce.type != 33) {
		return __LINE__;
	}

	if (ts_parse_actions[14].action.reduce.dynamic_precedence != 44) {
		return __LINE__;
	}

	return 0;
}
