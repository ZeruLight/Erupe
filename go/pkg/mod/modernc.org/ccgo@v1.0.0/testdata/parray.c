void f(char*){}

int main() {
	char a[10], b[20][30];
	//f(a);
	//f(b);
	f(b[2]);
	return 0;
}
