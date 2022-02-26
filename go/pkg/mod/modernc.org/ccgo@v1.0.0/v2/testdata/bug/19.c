struct s {
	int x;
};

int f(a, b, c) 
	char * b; // (a, b, c) vs (b, c, a)
	struct s c;
	int a;
{
	return a+c.x;
}

char * p;

int main() {
	struct s v;
	v.x = 314;
	return f(42, p, v) != 42+314;
}
