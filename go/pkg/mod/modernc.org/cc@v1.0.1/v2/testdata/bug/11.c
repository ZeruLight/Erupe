int __isnan(double) {}
int __isnanf(float) {}
int __isnanl(long double) {}

#define __typeof__ typeof
#define __dfp_expansion(__call, __fin, x) __fin
#define __mingw_choose_expr __builtin_choose_expr
#define __mingw_types_compatible_p(type1, type2) __builtin_types_compatible_p ( type1 , type2 )
#define isnan(x) __mingw_choose_expr ( __mingw_types_compatible_p ( __typeof__ ( x ) , double ) , __isnan ( x ) , __mingw_choose_expr ( __mingw_types_compatible_p ( __typeof__ ( x ) , float ) , __isnanf ( x ) , __mingw_choose_expr ( __mingw_types_compatible_p ( __typeof__ ( x ) , long double ) , __isnanl ( x ) , __dfp_expansion ( __isnan , ( __builtin_trap ( ) , x ) , x ) ) ) )

int main() {
	float f;
	double d;
	long double l;
	isnan(f);
	isnan(d);
	isnan(l);
}
