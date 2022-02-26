#include  <stdint.h>

int main() {
	int y;
	int8_t i8;
	uint8_t u8;
	int16_t i16;
	uint16_t u16;
	int32_t i32;
	uint32_t u32;
	int64_t i64;
	uint64_t u64;

	for (i8 = -4; i8 <= 4; i8++) {
		for (y = -128; y <= 127; y++) {
			printf("  signed 8 %i(%x) << %i(%x) = %i(%x)\n", i8, (uint8_t)i8, y, (unsigned)y, i8 << y, (uint8_t)(i8 << y));
		}
	}
	for (u8 = 0; u8 <= 4; u8++) {
		for (y = -128; y <= 127; y++) {
			printf("unsigned 8 %u(%x) << %i(%x) = %u(%x)\n", u8, u8, y, (unsigned)y, u8 << y, u8 << y);
		}
	}
	for (i8 = -4; i8 <= 4; i8++) {
		for (y = -128; y <= 127; y++) {
			printf("  signed 8 %i(%x) >> %i(%x) = %i(%x)\n", i8, (uint8_t)i8, y, (unsigned)y, i8 >> y, (uint8_t)(i8 >> y));
		}
	}
	for (u8 = 0; u8 <= 4; u8++) {
		for (y = -128; y <= 127; y++) {
			printf("unsigned 8 %u(%x) >> %i(%x) = %u(%x)\n", u8, u8, y, (unsigned)y, u8 >> y, u8 >> y);
		}
	}

	for (i16 = -4; i16 <= 4; i16++) {
		for (y = -128; y <= 127; y++) {
			printf("  signed 8 %i(%x) << %i(%x) = %i(%x)\n", i16, (uint16_t)i16, y, (unsigned)y, i16 << y, (uint16_t)(i16 << y));
		}
	}
	for (u16 = 0; u16 <= 4; u16++) {
		for (y = -128; y <= 127; y++) {
			printf("unsigned 8 %u(%x) << %i(%x) = %u(%x)\n", u16, u16, y, (unsigned)y, u16 << y, u16 << y);
		}
	}
	for (i16 = -4; i16 <= 4; i16++) {
		for (y = -128; y <= 127; y++) {
			printf("  signed 8 %i(%x) >> %i(%x) = %i(%x)\n", i16, (uint16_t)i16, y, (unsigned)y, i16 >> y, (uint16_t)(i16 >> y));
		}
	}
	for (u16 = 0; u16 <= 4; u16++) {
		for (y = -128; y <= 127; y++) {
			printf("unsigned 8 %u(%x) >> %i(%x) = %u(%x)\n", u16, u16, y, (unsigned)y, u16 >> y, u16 >> y);
		}
	}

	for (i32 = -4; i32 <= 4; i32++) {
		for (y = -128; y <= 127; y++) {
			printf("  signed 8 %i(%x) << %i(%x) = %i(%x)\n", i32, (uint32_t)i32, y, (unsigned)y, i32 << y, (uint32_t)(i32 << y));
		}
	}
	for (u32 = 0; u32 <= 4; u32++) {
		for (y = -128; y <= 127; y++) {
			printf("unsigned 8 %u(%x) << %i(%x) = %u(%x)\n", u32, u32, y, (unsigned)y, u32 << y, u32 << y);
		}
	}
	for (i32 = -4; i32 <= 4; i32++) {
		for (y = -128; y <= 127; y++) {
			printf("  signed 8 %i(%x) >> %i(%x) = %i(%x)\n", i16, (uint32_t)i16, y, (unsigned)y, i16 >> y, (uint32_t)(i16 >> y));
		}
	}
	for (u32 = 0; u32 <= 4; u32++) {
		for (y = -128; y <= 127; y++) {
			printf("unsigned 8 %u(%x) >> %i(%x) = %u(%x)\n", u32, u32, y, (unsigned)y, u32 >> y, u32 >> y);
		}
	}

	for (i64 = -4; i64 <= 4; i64++) {
		for (y = -128; y <= 127; y++) {
			printf("  signed 8 %i(%x) << %i(%x) = %i(%x)\n", i64, (uint64_t)i64, y, (unsigned)y, i64 << y, (uint64_t)(i64 << y));
		}
	}
	for (u64 = 0; u64 <= 4; u64++) {
		for (y = -128; y <= 127; y++) {
			printf("unsigned 8 %u(%x) << %i(%x) = %u(%x)\n", u64, u64, y, (unsigned)y, u64 << y, u64 << y);
		}
	}
	for (i64 = -4; i64 <= 4; i64++) {
		for (y = -128; y <= 127; y++) {
			printf("  signed 8 %i(%x) >> %i(%x) = %i(%x)\n", i16, (uint64_t)i16, y, (unsigned)y, i16 >> y, (uint64_t)(i16 >> y));
		}
	}
	for (u64 = 0; u64 <= 4; u64++) {
		for (y = -128; y <= 127; y++) {
			printf("unsigned 8 %u(%x) >> %i(%x) = %u(%x)\n", u64, u64, y, (unsigned)y, u64 >> y, u64 >> y);
		}
	}
}
