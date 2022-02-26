//  Copyright 2018 The CC Authors. All rights reserved.
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.

// +build ignore

#ifndef _BUILTIN_H_
#define _BUILTIN_H_

#include "predefined.h"

typedef void *__builtin_va_list;

typedef void *__FILE_TYPE__;

__SIZE_TYPE__ __builtin_strlen(char *__s);
__UINT64_TYPE__ __builtin_bswap64(__UINT64_TYPE__ x);
__builtin_va_list __builtin_va_copy();
char *__builtin_strchrnul(char *, int);
char *__builtin_strcpy(char *__dest, char *__src);
double _Complex __builtin_complex(double, double);
double __builtin_cimag(double _Complex);
double __builtin_copysign(double x, double y);
float _Complex __builtin_complexf(float, float);
float __builtin_cimagf(float _Complex);
int __builtin_abs(int j);
int __builtin_ffs(int i);
int __builtin_isprint(int);
int __builtin_memcmp(void *__s1, void *__s2, __SIZE_TYPE__ __n);
int __builtin_printf(char *__format, ...);
int __builtin_sprintf(char *__s, char *__format, ...);
int __builtin_strcmp(char *__s1, char *__s2);
int __signbit(double x);
int __signbitf(float x);
long double _Complex __builtin_complexl(double, double);
long double __builtin_cimagl(long double _Complex);
void *__GO2__(char*);
void *__builtin_alloca(__SIZE_TYPE__ __size);
void *__builtin_memcpy(void *dest, const void *src, __SIZE_TYPE__ n);
void *__builtin_memset(void *s, int c, __SIZE_TYPE__ n);
void __GO__(char*);
void __builtin_abort(void);
void __builtin_exit(int __status);
void __builtin_free(void*);
void __builtin_trap(void);
void __builtin_va_end(__builtin_va_list);
void __register_stdfiles(void *, void *, void *, void *);

#define __builtin_choose_expr(a, b, c) (a) ? (b) : (c)
#define __builtin_expect(exp, c) (exp)
#define __builtin_offsetof(st, m) ((__SIZE_TYPE__)(&((st *)0)->m))
#define __builtin_prefetch(addr, ...) (void)(addr)
#define __builtin_signbit(x) (sizeof(x) == sizeof(float) ? __signbitf(x) : sizeof (x) == sizeof(double) ? __signbit(x) : __signbitl(x))
#define __builtin_types_compatible_p(type1, type2) __builtin_types_compatible__((type1){}, (type2){})
#define __builtin_va_arg(ap, type) (type)ap
#ifdef __ccgo__
void __log(char*, ...);
#define __builtin_va_copy(dest, src) dest = __builtin_va_copy(src)
#define __builtin_va_end(ap) __builtin_va_end(ap)
#define __builtin_va_start(ap, arg) ap = __GO2__("X__builtin_va_start(tls, ap)")
#define __log(s, ...) __log("%s:%i.%s: " s, __FILE__, __LINE__, __func__, __VA_ARGS__)
#else
#define __builtin_va_copy(dest, src) dest = src
#define __builtin_va_end(ap) ap = (void*)0
#define __builtin_va_start(ap, arg) ap = (void*)-1
#endif
#define __complex__ _Complex
#define __const const
#define __extension__
#define __typeof__ typeof
#define __volatile volatile

#endif				/* _BUILTIN_H_ */
