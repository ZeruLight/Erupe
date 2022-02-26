set -e
rm -f log-ccgo
make clean || true
make distclean || true
./configure CC=ccgo \
	CFLAGS='--ccgo-full-paths --ccgo-struct-checks --ccgo-use-import exec.ErrNotFound,os.DevNull -D_GNU_SOURCE' \
	LDFLAGS='--warn-unresolved-libs --warn-go-build --ccgo-go --ccgo-import os,os/exec'
make binaries
make test
date

# all.tcl:	Total	25904	Passed	1530	Skipped	943	Failed	23431	# custom match fail unpatched
# all.tcl:	Total	25904	Passed	1530	Skipped	884	Failed	23490	# with -DTCL_MEM_DEBUG
# all.tcl:	Total	13453	Passed	12580	Skipped	831	Failed	42 	# removed -DTCL_MEM_DEBUG, trying musl memory allocator
# all.tcl:	Total	15687	Passed	14883	Skipped	798	Failed	6	# musl memory allocator with -DTCL_MEM_DEBUG
# all.tcl:	Total	25904	Passed	24963	Skipped	884	Failed	57	# Fixed vdso clock_gettime
# all.tcl:	Total	14925	Passed	13922	Skipped	909	Failed	94	# Removed -DTCL_MEM_DEBUG
# all.tcl:	Total	25959	Passed	24919	Skipped	944	Failed	96	# dtto
