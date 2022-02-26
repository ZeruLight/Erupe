struct s {
	char *s;
	int mxAlloc;
	int nChar;
} t;

int main() {
	f(&t);
}

char *f(struct s *p){
  if( p->s ){
    if( p ){
      return 0;
    }
  }
  return p->s;
}
