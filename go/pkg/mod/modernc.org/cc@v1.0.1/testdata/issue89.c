struct a
{
  unsigned f : 3;
};

void h()
{
  struct a a;
  g(a.f);
}

int g(unsigned);
