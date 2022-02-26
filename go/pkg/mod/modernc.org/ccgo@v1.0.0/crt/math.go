// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crt // import "modernc.org/ccgo/crt"

import (
	"math"
)

func Xacos(tls *TLS, x float64) float64        { return math.Acos(x) }
func Xasin(tls *TLS, x float64) float64        { return math.Asin(x) }
func Xatan(tls *TLS, x float64) float64        { return math.Atan(x) }
func Xceil(tls *TLS, x float64) float64        { return math.Ceil(x) }
func Xcopysign(tls *TLS, x, y float64) float64 { return X__builtin_copysign(tls, x, y) }
func Xcos(tls *TLS, x float64) float64         { return math.Cos(x) }
func Xcosh(tls *TLS, x float64) float64        { return math.Cosh(x) }
func Xexp(tls *TLS, x float64) float64         { return math.Exp(x) }
func Xfabs(tls *TLS, x float64) float64        { return math.Abs(x) }
func Xfloor(tls *TLS, x float64) float64       { return math.Floor(x) }
func Xlog(tls *TLS, x float64) float64         { return math.Log(x) }
func Xlog10(tls *TLS, x float64) float64       { return math.Log10(x) }
func Xpow(tls *TLS, x, y float64) float64      { return math.Pow(x, y) }
func Xsin(tls *TLS, x float64) float64         { return math.Sin(x) }
func Xsinh(tls *TLS, x float64) float64        { return math.Sinh(x) }
func Xsqrt(tls *TLS, x float64) float64        { return math.Sqrt(x) }
func Xtan(tls *TLS, x float64) float64         { return math.Tan(x) }
func Xtanh(tls *TLS, x float64) float64        { return math.Tanh(x) }

// double round(double x);
func Xround(tls *TLS, x float64) float64 {
	switch {
	case x < 0:
		return math.Ceil(x - 0.5)
	case x > 0:
		return math.Floor(x + 0.5)
	}
	return x
}

// int __signbit(double x);
func X__signbit(tls *TLS, x float64) int32 {
	if math.Signbit(x) {
		return 1
	}

	return 0
}

// int __signbitf(float x);
func X__signbitf(tls *TLS, x float32) int32 {
	if math.Signbit(float64(x)) {
		return 1
	}

	return 0
}

func X__builtin_copysign(tls *TLS, x, y float64) float64 { return math.Copysign(x, y) }
