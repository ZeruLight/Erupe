typedef unsigned int u_int32_t __attribute__ ((__mode__ (__SI__)));
typedef unsigned short u_int16_t __attribute__ ((__mode__ (__SI__)));
typedef unsigned short int __uint16_t;

enum {
    kIsInvisible = 0x4000,
};

typedef struct finderinfo {
    u_int16_t fdFlags;
} __attribute__ ((__packed__)) finderinfo;

typedef struct fileinfobuf {
    u_int32_t info_length;
    u_int32_t data[8];
} fileinfobuf;

void main() {
  fileinfobuf finfo;
  finderinfo *finder = (finderinfo *) &finfo.data;

  printf("%d\n", kIsInvisible);
  printf("%d\n", (((__uint16_t)((((__uint16_t)(kIsInvisible) & 0xff00) >> 8) |          (((__uint16_t)(kIsInvisible) & 0x00ff) << 8)))));
  printf("%d\n", finder->fdFlags);
  printf("%d\n", ~(((__uint16_t)((((__uint16_t)(kIsInvisible) & 0xff00) >> 8) |                 (((__uint16_t)(kIsInvisible) & 0x00ff) << 8)))));
  printf("%d\n", 0 & ~(((__uint16_t)((((__uint16_t)(kIsInvisible) & 0xff00) >> 8) |             (((__uint16_t)(kIsInvisible) & 0x00ff) << 8)))));

  finder->fdFlags &= ~(((__uint16_t)((((__uint16_t)(kIsInvisible) & 0xff00) >> 8) |             (((__uint16_t)(kIsInvisible) & 0x00ff) << 8))));

  printf("%d\n", finder->fdFlags);
}
