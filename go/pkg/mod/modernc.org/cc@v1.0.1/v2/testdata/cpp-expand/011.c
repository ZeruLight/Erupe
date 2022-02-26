#if 0 != (0 && (0/0))
   FAIL
#endif

#if 1 != (-1 || (0/0))
   FAIL
#endif

#if 3 != (-1 ? 3 : (0/0))
   FAIL
#endif

int
main()
{
	return 0;
}
