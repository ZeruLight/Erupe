int atomic_fetch_add(unsigned int *ptr, int v) {
	return *ptr+v;
}


int atomic_add_int(int *ptr, int v)
{
    return atomic_fetch_add((_Atomic(unsigned int) *)ptr, v) + v;
}

