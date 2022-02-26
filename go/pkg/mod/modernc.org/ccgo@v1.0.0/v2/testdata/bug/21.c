#include <assert.h>

int f() {

//	return (
//		(union { _Complex float __z; float __xy[2]; }) {
//			.__xy = {
//				((float)1.57079632679489661923 - ((float)(z))),
//				(
//					-(+(union { _Complex float __z; float __xy[2]; }){
//						(_Complex float)(z)
//					}.__xy[1])
//				)
//			}
//		}.__z
//	);

	return (union { int i; int a[2]; }) {
			.a = {3, 4},
		}.a[1];
}

int main() {
	assert(f() == 4);
}
