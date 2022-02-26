union U1 {
   signed f0 : 25;
   unsigned  f1;
};

static union U1 g_60 = {0x6E9B1CC8L};

int main() {
	__builtin_printf("%i\n", g_60.f0);
}
