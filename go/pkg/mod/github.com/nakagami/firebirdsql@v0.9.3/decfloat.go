/*******************************************************************************
The MIT License (MIT)

Copyright (c) 2019 Hajime Nakagami

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*******************************************************************************/

package firebirdsql

import (
	"github.com/shopspring/decimal"
	"math"
	"math/big"
)

func dpdBitToInt(dpd uint, mask uint) int {
	if (dpd & mask) != 0 {
		return 1
	}
	return 0
}

func dpdToInt(dpd uint) int64 {
	// Convert DPD encodined value to int (0-999)
	// dpd: DPD encoded value. 10bit unsigned int

	b := make([]int, 10)
	b[9] = dpdBitToInt(dpd, 0x0200)
	b[8] = dpdBitToInt(dpd, 0x0100)
	b[7] = dpdBitToInt(dpd, 0x0080)
	b[6] = dpdBitToInt(dpd, 0x0040)
	b[5] = dpdBitToInt(dpd, 0x0020)
	b[4] = dpdBitToInt(dpd, 0x0010)
	b[3] = dpdBitToInt(dpd, 0x0008)
	b[2] = dpdBitToInt(dpd, 0x0004)
	b[1] = dpdBitToInt(dpd, 0x0002)
	b[0] = dpdBitToInt(dpd, 0x0001)

	d := make([]int, 3)
	if b[3] == 0 {
		d[2] = b[9]*4 + b[8]*2 + b[7]
		d[1] = b[6]*4 + b[5]*2 + b[4]
		d[0] = b[2]*4 + b[1]*2 + b[0]
	} else if b[3] == 1 && b[2] == 0 && b[1] == 0 {
		d[2] = b[9]*4 + b[8]*2 + b[7]
		d[1] = b[6]*4 + b[5]*2 + b[4]
		d[0] = 8 + b[0]
	} else if b[3] == 1 && b[2] == 0 && b[1] == 1 {
		d[2] = b[9]*4 + b[8]*2 + b[7]
		d[1] = 8 + b[4]
		d[0] = b[6]*4 + b[5]*2 + b[0]
	} else if b[3] == 1 && b[2] == 1 && b[1] == 0 {
		d[2] = 8 + b[7]
		d[1] = b[6]*4 + b[5]*2 + b[4]
		d[0] = b[9]*4 + b[8]*2 + b[0]
	} else if b[6] == 0 && b[5] == 0 && b[3] == 1 && b[2] == 1 && b[1] == 1 {
		d[2] = 8 + b[7]
		d[1] = 8 + b[4]
		d[0] = b[9]*4 + b[8]*2 + b[0]
	} else if b[6] == 0 && b[5] == 1 && b[3] == 1 && b[2] == 1 && b[1] == 1 {
		d[2] = 8 + b[7]
		d[1] = b[9]*4 + b[8]*2 + b[4]
		d[0] = 8 + b[0]
	} else if b[6] == 1 && b[5] == 0 && b[3] == 1 && b[2] == 1 && b[1] == 1 {
		d[2] = b[9]*4 + b[8]*2 + b[7]
		d[1] = 8 + b[4]
		d[0] = 8 + b[0]
	} else if b[6] == 1 && b[5] == 1 && b[3] == 1 && b[2] == 1 && b[1] == 1 {
		d[2] = 8 + b[7]
		d[1] = 8 + b[4]
		d[0] = 8 + b[0]
	} else {
		panic("Invalid DPD encoding")
	}

	return int64(d[2])*100 + int64(d[1])*10 + int64(d[0])
}

func calcSignificand(prefix int64, dpdBits *big.Int, numBits int) *big.Int {
	// prefix: High bits integer value
	// dpdBits: dpd encoded bits
	// numBits: bit length of dpd_bits
	// https://en.wikipedia.org/wiki/Decimal128_floating-point_format#Densely_packed_decimal_significand_field
	numSegments := numBits / 10
	segments := make([]uint, numSegments)
	bi1024 := big.NewInt(1024)

	for i := 0; i < numSegments; i++ {
		var work big.Int
		work.Add(&work, dpdBits)
		segments[numSegments-i-1] = uint(work.Mod(&work, bi1024).Int64())
		dpdBits.Rsh(dpdBits, 10)
	}

	v := big.NewInt(prefix)
	bi1000 := big.NewInt(1000)
	for _, dpd := range segments {
		v.Mul(v, bi1000)
		v.Add(v, big.NewInt(dpdToInt(dpd)))
	}

	return v
}

func decimal128ToSignDigitsExponent(b []byte) (v *decimal.Decimal, sign int, digits *big.Int, exponent int32) {
	// https://en.wikipedia.org/wiki/Decimal128_floating-point_format

	var prefix int64
	if (b[0] & 0x80) == 0x80 {
		sign = 1
	}
	cf := (uint32(b[0]&0x7f) << 10) + uint32(b[1]<<2) + uint32(b[2]>>6)
	if (cf & 0x1F000) == 0x1F000 {
		var d decimal.Decimal
		if sign == 1 {
			// Is there -NaN ?
			d = decimal.NewFromFloat(math.NaN())
		} else {
			d = decimal.NewFromFloat(math.NaN())
		}
		v = &d
		return
	} else if (cf & 0x1F000) == 0x1E000 {
		var d decimal.Decimal
		if sign == 1 {
			d = decimal.NewFromFloat(math.Inf(-1))
		} else {
			d = decimal.NewFromFloat(math.Inf(1))
		}
		v = &d
		return
	} else if (cf & 0x18000) == 0x00000 {
		exponent = int32(0x0000 + (cf & 0x00fff))
		prefix = int64((cf >> 12) & 0x07)
	} else if (cf & 0x18000) == 0x08000 {
		exponent = int32(0x1000 + (cf & 0x00fff))
		prefix = int64((cf >> 12) & 0x07)
	} else if (cf & 0x18000) == 0x10000 {
		exponent = int32(0x2000 + (cf & 0x00fff))
		prefix = int64((cf >> 12) & 0x07)
	} else if (cf & 0x1e000) == 0x18000 {
		exponent = int32(0x0000 + (cf & 0x00fff))
		prefix = int64(8 + (cf>>12)&0x01)
	} else if (cf & 0x1e000) == 0x1a000 {
		exponent = int32(0x1000 + (cf & 0x00fff))
		prefix = int64(8 + (cf>>12)&0x01)
	} else if (cf & 0x1e000) == 0x1c000 {
		exponent = int32(0x2000 + (cf & 0x00fff))
		prefix = int64(8 + (cf>>12)&0x01)
	} else {
		panic("decimal128 value error")
	}
	exponent -= 6176

	dpdBits := bytesToBigInt(b)
	mask := bigIntFromHexString("3fffffffffffffffffffffffffff")
	dpdBits.And(dpdBits, mask)
	digits = calcSignificand(prefix, dpdBits, 110)

	return
}

func decimalFixedToDecimal(b []byte, scale int32) decimal.Decimal {
	v, sign, digits, _ := decimal128ToSignDigitsExponent(b)
	if v != nil {
		return *v
	}
	if sign != 0 {
		digits.Mul(digits, big.NewInt(-1))
	}
	return decimal.NewFromBigInt(digits, scale)
}

func decimal64ToDecimal(b []byte) decimal.Decimal {
	// https://en.wikipedia.org/wiki/Decimal64_floating-point_format
	var prefix int64
	var sign int
	if (b[0] & 0x80) == 0x80 {
		sign = 1
	}
	cf := (uint32(b[0]) >> 2) & 0x1f
	exponent := ((int32(b[0]) & 3) << 6) + ((int32(b[1]) >> 2) & 0x3f)

	dpdBits := bytesToBigInt(b)
	mask := bigIntFromHexString("3ffffffffffff")
	dpdBits.And(dpdBits, mask)

	if cf == 0x1f {
		if sign == 1 {
			// Is there -NaN ?
			return decimal.NewFromFloat(math.NaN())
		}
		return decimal.NewFromFloat(math.NaN())
	} else if cf == 0x1e {
		if sign == 1 {
			return decimal.NewFromFloat(math.Inf(-1))
		}
		return decimal.NewFromFloat(math.Inf(1))
	} else if (cf & 0x18) == 0x00 {
		exponent = 0x000 + exponent
		prefix = int64(cf & 0x07)
	} else if (cf & 0x18) == 0x08 {
		exponent = 0x100 + exponent
		prefix = int64(cf & 0x07)
	} else if (cf & 0x18) == 0x10 {
		exponent = 0x200 + exponent
		prefix = int64(cf & 0x07)
	} else if (cf & 0x1e) == 0x18 {
		exponent = 0x000 + exponent
		prefix = int64(8 + cf&1)
	} else if (cf & 0x1e) == 0x1a {
		exponent = 0x100 + exponent
		prefix = int64(8 + cf&1)
	} else if (cf & 0x1e) == 0x1c {
		exponent = 0x200 + exponent
		prefix = int64(8 + cf&1)
	} else {
		panic("decimal64 value error")
	}
	digits := calcSignificand(prefix, dpdBits, 50)
	exponent -= 398

	if sign != 0 {
		digits.Mul(digits, big.NewInt(-1))
	}
	return decimal.NewFromBigInt(digits, exponent)
}

func decimal128ToDecimal(b []byte) decimal.Decimal {
	// https://en.wikipedia.org/wiki/Decimal64_floating-point_format
	v, sign, digits, exponent := decimal128ToSignDigitsExponent(b)
	if v != nil {
		return *v
	}
	if sign != 0 {
		digits.Mul(digits, big.NewInt(-1))
	}
	return decimal.NewFromBigInt(digits, exponent)
}
