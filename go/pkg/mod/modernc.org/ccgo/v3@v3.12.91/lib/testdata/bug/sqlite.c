struct item {
	char *zColl;
};

struct item item = {"foo"};

void f1(struct item item) {
	__builtin_printf("%i: %s\n", __LINE__, item.zColl);
}

void f2(struct item *item) {
	__builtin_printf("%i: %s\n", __LINE__, item->zColl);
}

struct tab {
	struct item aCol[3];
};

struct tab tab = {{{"foo"}, {"bar"}, {"baz"}}};

int j = 1;

void f3(struct tab tab) {
	__builtin_printf("%i: tab.aCol[j].zColl %s\n", __LINE__, tab.aCol[j].zColl);
}

void f4(struct tab *ptab) {
	__builtin_printf("%i: ptab->aCol[j].zColl %s\n", __LINE__, ptab->aCol[j].zColl);
}

struct y {
	struct tab *ptab;
};

struct y y = {&tab};

void f5(struct y y) {
	__builtin_printf("%i: y.ptab->aCol[j].zColl %s\n", __LINE__, y.ptab->aCol[j].zColl);
}

void f6(struct y *y) {
	__builtin_printf("%i: y->ptab->aCol[j].zColl %s\n", __LINE__, y->ptab->aCol[j].zColl);
}

struct p {
	struct y y;
};

struct p p;

void f7(struct p p) {
	__builtin_printf("%i: p.y.ptab->aCol[j].zColl %s\n", __LINE__, p.y.ptab->aCol[j].zColl);
}

void f8(struct p *p) {
	__builtin_printf("%i: p->y.ptab->aCol[j].zColl %s\n", __LINE__, p->y.ptab->aCol[j].zColl);
}

// const char *zColl = p->y.pTab->aCol[j].zColl;

int main() {
	p.y = y;
	f1(item);
	f2(&item);
	f3(tab);
	f4(&tab);
	f5(y);
	f6(&y);
	f7(p);
	f8(&p);
}
