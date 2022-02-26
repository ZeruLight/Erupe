#define getVarint32(A,B)  \
  (u8)((*(A)<(u8)0x80)?((B)=(u32)*(A)),1:sqlite3GetVarint32((A),(u32 *)&(B)))

typedef unsigned char u8;
typedef unsigned u32;

struct {
	char *z;
} m;

static u8 sqlite3GetVarint32(const unsigned char *p, u32 *v){}

int main() {
	u32 szHdr;
	m.z = "foo";
	(void)getVarint32((u8*)m.z, szHdr);
	if (szHdr != 'f') {
		return 1;
	}
}
