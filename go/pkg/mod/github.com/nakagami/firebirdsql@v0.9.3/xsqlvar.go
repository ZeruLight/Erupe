/*******************************************************************************
The MIT License (MIT)

Copyright (c) 2013-2020 Hajime Nakagami

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
	"bytes"
	"encoding/binary"
	"github.com/shopspring/decimal"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"math"
	"math/big"
	"reflect"
	"time"
)

const (
	SQL_TYPE_TEXT         = 452
	SQL_TYPE_VARYING      = 448
	SQL_TYPE_SHORT        = 500
	SQL_TYPE_LONG         = 496
	SQL_TYPE_FLOAT        = 482
	SQL_TYPE_DOUBLE       = 480
	SQL_TYPE_D_FLOAT      = 530
	SQL_TYPE_TIMESTAMP    = 510
	SQL_TYPE_BLOB         = 520
	SQL_TYPE_ARRAY        = 540
	SQL_TYPE_QUAD         = 550
	SQL_TYPE_TIME         = 560
	SQL_TYPE_DATE         = 570
	SQL_TYPE_INT64        = 580
	SQL_TYPE_INT128       = 32752
	SQL_TYPE_TIMESTAMP_TZ = 32754
	SQL_TYPE_TIME_TZ      = 32756
	SQL_TYPE_DEC_FIXED    = 32758
	SQL_TYPE_DEC64        = 32760
	SQL_TYPE_DEC128       = 32762
	SQL_TYPE_BOOLEAN      = 32764
	SQL_TYPE_NULL         = 32766
)

var xsqlvarTypeLength = map[int]int{
	SQL_TYPE_TEXT:         -1,
	SQL_TYPE_VARYING:      -1,
	SQL_TYPE_SHORT:        4,
	SQL_TYPE_LONG:         4,
	SQL_TYPE_FLOAT:        4,
	SQL_TYPE_TIME:         4,
	SQL_TYPE_DATE:         4,
	SQL_TYPE_DOUBLE:       8,
	SQL_TYPE_TIMESTAMP:    8,
	SQL_TYPE_BLOB:         8,
	SQL_TYPE_ARRAY:        8,
	SQL_TYPE_QUAD:         8,
	SQL_TYPE_INT64:        8,
	SQL_TYPE_INT128:       16,
	SQL_TYPE_TIMESTAMP_TZ: 10,
	SQL_TYPE_TIME_TZ:      6,
	SQL_TYPE_DEC64:        8,
	SQL_TYPE_DEC128:       16,
	SQL_TYPE_DEC_FIXED:    16,
	SQL_TYPE_BOOLEAN:      1,
}

var xsqlvarTypeDisplayLength = map[int]int{
	SQL_TYPE_TEXT:         -1,
	SQL_TYPE_VARYING:      -1,
	SQL_TYPE_SHORT:        6,
	SQL_TYPE_LONG:         11,
	SQL_TYPE_FLOAT:        17,
	SQL_TYPE_TIME:         11,
	SQL_TYPE_DATE:         10,
	SQL_TYPE_DOUBLE:       17,
	SQL_TYPE_TIMESTAMP:    22,
	SQL_TYPE_BLOB:         0,
	SQL_TYPE_ARRAY:        -1,
	SQL_TYPE_QUAD:         20,
	SQL_TYPE_INT64:        20,
	SQL_TYPE_INT128:       20,
	SQL_TYPE_TIMESTAMP_TZ: 28,
	SQL_TYPE_TIME_TZ:      17,
	SQL_TYPE_DEC64:        16,
	SQL_TYPE_DEC128:       34,
	SQL_TYPE_DEC_FIXED:    34,
	SQL_TYPE_BOOLEAN:      5,
}

var xsqlvarTypeName = map[int]string{
	SQL_TYPE_TEXT:         "TEXT",
	SQL_TYPE_VARYING:      "VARYING",
	SQL_TYPE_SHORT:        "SHORT",
	SQL_TYPE_LONG:         "LONG",
	SQL_TYPE_FLOAT:        "FLOAT",
	SQL_TYPE_TIME:         "TIME",
	SQL_TYPE_DATE:         "DATE",
	SQL_TYPE_DOUBLE:       "DOUBLE",
	SQL_TYPE_TIMESTAMP:    "TIMESTAMP",
	SQL_TYPE_BLOB:         "BLOB",
	SQL_TYPE_ARRAY:        "ARRAY",
	SQL_TYPE_QUAD:         "QUAD",
	SQL_TYPE_INT64:        "INT64",
	SQL_TYPE_INT128:       "INT128",
	SQL_TYPE_TIMESTAMP_TZ: "TIMESTAMP WITH TIMEZONE",
	SQL_TYPE_TIME_TZ:      "TIME WITH TIMEZONE",
	SQL_TYPE_DEC64:        "DECFLOAT(16)",
	SQL_TYPE_DEC128:       "DECFLOAT(34)",
	SQL_TYPE_DEC_FIXED:    "DECFIXED",
	SQL_TYPE_BOOLEAN:      "BOOLEAN",
}

type xSQLVAR struct {
	sqltype    int
	sqlscale   int
	sqlsubtype int
	sqllen     int
	null_ok    bool
	fieldname  string
	relname    string
	ownname    string
	aliasname  string
}

func (x *xSQLVAR) ioLength() int {
	if x.sqltype == SQL_TYPE_TEXT {
		return x.sqllen
	}
	return xsqlvarTypeLength[x.sqltype]
}

func (x *xSQLVAR) displayLength() int {
	if x.sqltype == SQL_TYPE_TEXT {
		return x.sqllen
	}
	return xsqlvarTypeDisplayLength[x.sqltype]
}

func (x *xSQLVAR) hasPrecisionScale() bool {
	return (x.sqltype == SQL_TYPE_SHORT || x.sqltype == SQL_TYPE_LONG || x.sqltype == SQL_TYPE_QUAD || x.sqltype == SQL_TYPE_INT64 || x.sqltype == SQL_TYPE_INT128 || x.sqltype == SQL_TYPE_DEC64 || x.sqltype == SQL_TYPE_DEC128 || x.sqltype == SQL_TYPE_DEC_FIXED) && x.sqlscale != 0
}

func (x *xSQLVAR) typename() string {
	return xsqlvarTypeName[x.sqltype]
}

func (x *xSQLVAR) scantype() reflect.Type {
	switch x.sqltype {
	case SQL_TYPE_TEXT:
		return reflect.TypeOf("")
	case SQL_TYPE_VARYING:
		return reflect.TypeOf("")
	case SQL_TYPE_SHORT:
		if x.sqlscale != 0 {
			return reflect.TypeOf(decimal.Decimal{})
		}
		return reflect.TypeOf(int16(0))
	case SQL_TYPE_LONG:
		if x.sqlscale != 0 {
			return reflect.TypeOf(decimal.Decimal{})
		}
		return reflect.TypeOf(int32(0))
	case SQL_TYPE_INT64:
		if x.sqlscale != 0 {
			return reflect.TypeOf(decimal.Decimal{})
		}
		return reflect.TypeOf(int64(0))
	case SQL_TYPE_INT128:
		return reflect.TypeOf(big.Int{})
	case SQL_TYPE_DATE:
		return reflect.TypeOf(time.Time{})
	case SQL_TYPE_TIME:
		return reflect.TypeOf(time.Time{})
	case SQL_TYPE_TIMESTAMP:
		return reflect.TypeOf(time.Time{})
	case SQL_TYPE_FLOAT:
		return reflect.TypeOf(float32(0))
	case SQL_TYPE_DOUBLE:
		return reflect.TypeOf(float64(0))
	case SQL_TYPE_BOOLEAN:
		return reflect.TypeOf(false)
	case SQL_TYPE_BLOB:
		return reflect.TypeOf([]byte{})
	case SQL_TYPE_TIMESTAMP_TZ:
		return reflect.TypeOf(time.Time{})
	case SQL_TYPE_TIME_TZ:
		return reflect.TypeOf(time.Time{})
	case SQL_TYPE_DEC64:
		return reflect.TypeOf(decimal.Decimal{})
	case SQL_TYPE_DEC128:
		return reflect.TypeOf(decimal.Decimal{})
	case SQL_TYPE_DEC_FIXED:
		return reflect.TypeOf(decimal.Decimal{})
	}
	return reflect.TypeOf(nil)
}

func (x *xSQLVAR) _parseTimezone(raw_value []byte) *time.Location {
	timezone := getTimezoneNameByID(int(bytes_to_buint16(raw_value)))
	tz, _ := time.LoadLocation(timezone)
	return tz
}

func (x *xSQLVAR) _parseDate(raw_value []byte) (int, int, int) {
	nday := int(bytes_to_bint32(raw_value)) + 678882
	century := (4*nday - 1) / 146097
	nday = 4*nday - 1 - 146097*century
	day := nday / 4

	nday = (4*day + 3) / 1461
	day = 4*day + 3 - 1461*nday
	day = (day + 4) / 4

	month := (5*day - 3) / 153
	day = 5*day - 3 - 153*month
	day = (day + 5) / 5
	year := 100*century + nday
	if month < 10 {
		month += 3
	} else {
		month -= 9
		year++
	}
	return year, month, day
}

func (x *xSQLVAR) _parseTime(raw_value []byte) (int, int, int, int) {
	n := int(bytes_to_bint32(raw_value))
	s := n / 10000
	m := s / 60
	h := m / 60
	m = m % 60
	s = s % 60
	return h, m, s, (n % 10000) * 100000
}

func (x *xSQLVAR) parseDate(raw_value []byte, timezone string) time.Time {
	tz := time.Local
	if timezone != "" {
		tz, _ = time.LoadLocation(timezone)
	}
	year, month, day := x._parseDate(raw_value)
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, tz)
}

func (x *xSQLVAR) parseTime(raw_value []byte, timezone string) time.Time {
	tz := time.Local
	if timezone != "" {
		tz, _ = time.LoadLocation(timezone)
	}
	h, m, s, n := x._parseTime(raw_value)
	now := time.Now()
	zone, offset := time.Date(now.Year(), now.Month(), now.Day(), h, m, s, n, tz).Zone()
	return time.Date(0, time.Month(1), 1, h, m, s, n, time.FixedZone(zone, offset))
}

func (x *xSQLVAR) parseTimestamp(raw_value []byte, timezone string) time.Time {
	tz := time.Local
	if timezone != "" {
		tz, _ = time.LoadLocation(timezone)
	}

	year, month, day := x._parseDate(raw_value[:4])
	h, m, s, n := x._parseTime(raw_value[4:8])
	return time.Date(year, time.Month(month), day, h, m, s, n, tz)
}

func (x *xSQLVAR) parseTimeTz(raw_value []byte) time.Time {
	h, m, s, n := x._parseTime(raw_value[:4])
	tz := x._parseTimezone(raw_value[4:6])
	loc := x._parseTimezone(raw_value[6:8])
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), h, m, s, n, tz).In(loc)
	zone, offset := t.Zone()
	return time.Date(0, time.Month(1), 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.FixedZone(zone, offset))
}

func (x *xSQLVAR) parseTimestampTz(raw_value []byte) time.Time {
	year, month, day := x._parseDate(raw_value[:4])
	h, m, s, n := x._parseTime(raw_value[4:8])
	tz := x._parseTimezone(raw_value[8:10])
	offset := x._parseTimezone(raw_value[10:12])
	return time.Date(year, time.Month(month), day, h, m, s, n, tz).In(offset)
}

func (x *xSQLVAR) parseString(raw_value []byte, charset string) interface{} {
	if x.sqlsubtype == 1 { // OCTETS
		return raw_value
	}
	if x.sqlsubtype == 0 {
		switch charset {
		case "OCTETS":
			return raw_value
		case "UNICODE_FSS", "UTF8":
			return bytes.NewBuffer(raw_value).String()
		case "SJIS_0208":
			dec := japanese.ShiftJIS.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "EUCJ_0208":
			dec := japanese.EUCJP.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "ISO8859_1":
			dec := charmap.ISO8859_1.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "ISO8859_2":
			dec := charmap.ISO8859_2.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "ISO8859_3":
			dec := charmap.ISO8859_3.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "ISO8859_4":
			dec := charmap.ISO8859_5.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "ISO8859_5":
			dec := charmap.ISO8859_5.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "ISO8859_6":
			dec := charmap.ISO8859_6.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "ISO8859_7":
			dec := charmap.ISO8859_7.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "ISO8859_8":
			dec := charmap.ISO8859_8.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "ISO8859_9":
			dec := charmap.ISO8859_9.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "ISO8859_13":
			dec := charmap.ISO8859_13.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "KSC_5601":
			dec := korean.EUCKR.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "WIN1250":
			dec := charmap.Windows1250.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "WIN1251":
			dec := charmap.Windows1251.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "WIN1252":
			dec := charmap.Windows1252.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "WIN1253":
			dec := charmap.Windows1252.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "WIN1254":
			dec := charmap.Windows1252.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "BIG_5":
			dec := traditionalchinese.Big5.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "GB_2312":
			dec := simplifiedchinese.HZGB2312.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "WIN1255":
			dec := charmap.Windows1255.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "WIN1256":
			dec := charmap.Windows1256.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "WIN1257":
			dec := charmap.Windows1257.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "KOI8R":
			dec := charmap.KOI8R.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "KOI8U":
			dec := charmap.KOI8U.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		case "WIN1258":
			dec := charmap.Windows1258.NewDecoder()
			v, _ := dec.Bytes(raw_value)
			return string(v)
		default:
			return bytes.NewBuffer(raw_value).String()
		}
	}

	return raw_value
}

func (x *xSQLVAR) value(raw_value []byte, timezone string, charset string) (v interface{}, err error) {
	switch x.sqltype {
	case SQL_TYPE_TEXT:
		if x.sqlsubtype == 1 { // OCTETS
			v = raw_value
		} else {
			v = x.parseString(raw_value, charset)
		}
	case SQL_TYPE_VARYING:
		if x.sqlsubtype == 1 { // OCTETS
			v = raw_value
		} else {
			v = x.parseString(raw_value, charset)
		}
	case SQL_TYPE_SHORT:
		i16 := int16(bytes_to_bint32(raw_value))
		if x.sqlscale > 0 {
			v = int64(i16) * int64(math.Pow10(x.sqlscale))
		} else if x.sqlscale < 0 {
			v = decimal.New(int64(i16), int32(x.sqlscale))
		} else {
			v = i16
		}
	case SQL_TYPE_LONG:
		i32 := bytes_to_bint32(raw_value)
		if x.sqlscale > 0 {
			v = int64(i32) * int64(math.Pow10(x.sqlscale))
		} else if x.sqlscale < 0 {
			v = decimal.New(int64(i32), int32(x.sqlscale))
		} else {
			v = i32
		}
	case SQL_TYPE_INT64:
		i64 := bytes_to_bint64(raw_value)
		if x.sqlscale > 0 {
			v = i64 * int64(math.Pow10(x.sqlscale))
		} else if x.sqlscale < 0 {
			v = decimal.New(int64(i64), int32(x.sqlscale))
		} else {
			v = i64
		}
	case SQL_TYPE_INT128:
		var isNegative bool

		// when raw_value[0] is > 127, then subtract 255 in every index
		if raw_value[0] > 127 {
			for i := range raw_value {
				if raw_value[i] < 255 {
					raw_value[i] = 255 - raw_value[i]
				} else {
					raw_value[i] -= 255
				}
			}
			isNegative = true
		}

		// reverse
		for i, j := 0, len(raw_value)-1; i < j; i, j = i+1, j-1 {
			raw_value[i], raw_value[j] = raw_value[j], raw_value[i]
		}

		// variable to return
		var x = new(big.Int)

		for i := len(raw_value) - 1; i >= 0; i-- {
			if raw_value[i] == 0 {
				continue
			}

			// get the 2^(i*8) in big.Float
			var t = new(big.Float).SetFloat64(math.Pow(2, float64(i*8)))

			// convert the float to int
			var xx *big.Int
			xx, _ = t.Int(xx)

			// mul with the value in raw_value
			xx.Mul(xx, big.NewInt(int64(raw_value[i])))

			// add to x
			x.Add(x, xx)
		}

		// when negative, add 1 and mul -1
		if isNegative {
			x.Add(x, big.NewInt(1))
			x.Mul(x, big.NewInt(-1))
		}
		v = x
	case SQL_TYPE_DATE:
		v = x.parseDate(raw_value, timezone)
	case SQL_TYPE_TIME:
		v = x.parseTime(raw_value, timezone)
	case SQL_TYPE_TIMESTAMP:
		v = x.parseTimestamp(raw_value, timezone)
	case SQL_TYPE_TIME_TZ:
		v = x.parseTimeTz(raw_value)
	case SQL_TYPE_TIMESTAMP_TZ:
		v = x.parseTimestampTz(raw_value)
	case SQL_TYPE_FLOAT:
		var f32 float32
		b := bytes.NewReader(raw_value)
		err = binary.Read(b, binary.BigEndian, &f32)
		v = f32
	case SQL_TYPE_DOUBLE:
		b := bytes.NewReader(raw_value)
		var f64 float64
		err = binary.Read(b, binary.BigEndian, &f64)
		v = f64
	case SQL_TYPE_BOOLEAN:
		v = raw_value[0] != 0
	case SQL_TYPE_BLOB:
		v = raw_value
	case SQL_TYPE_DEC_FIXED:
		v = decimalFixedToDecimal(raw_value, int32(x.sqlscale))
	case SQL_TYPE_DEC64:
		v = decimal64ToDecimal(raw_value)
	case SQL_TYPE_DEC128:
		v = decimal128ToDecimal(raw_value)
	}
	return
}

func calcBlr(xsqlda []xSQLVAR) []byte {
	// Calculate  BLR from XSQLVAR array.
	ln := len(xsqlda) * 2
	blr := make([]byte, (ln*4)+8)
	blr[0] = 5
	blr[1] = 2
	blr[2] = 4
	blr[3] = 0
	blr[4] = byte(ln & 255)
	blr[5] = byte(ln >> 8)
	n := 6

	for _, x := range xsqlda {
		sqlscale := x.sqlscale
		if sqlscale < 0 {
			sqlscale += 256
		}
		switch x.sqltype {
		case SQL_TYPE_VARYING:
			blr[n] = 37
			blr[n+1] = byte(x.sqllen & 255)
			blr[n+2] = byte(x.sqllen >> 8)
			n += 3
		case SQL_TYPE_TEXT:
			blr[n] = 14
			blr[n+1] = byte(x.sqllen & 255)
			blr[n+2] = byte(x.sqllen >> 8)
			n += 3
		case SQL_TYPE_LONG:
			blr[n] = 8
			blr[n+1] = byte(sqlscale)
			n += 2
		case SQL_TYPE_SHORT:
			blr[n] = 7
			blr[n+1] = byte(sqlscale)
			n += 2
		case SQL_TYPE_INT64:
			blr[n] = 16
			blr[n+1] = byte(sqlscale)
			n += 2
		case SQL_TYPE_INT128:
			blr[n] = 26
			blr[n+1] = byte(sqlscale)
			n += 2
		case SQL_TYPE_QUAD:
			blr[n] = 9
			blr[n+1] = byte(sqlscale)
			n += 2
		case SQL_TYPE_DEC_FIXED: // OBSOLATED
			blr[n] = 26
			blr[n+1] = byte(sqlscale)
			n += 2
		case SQL_TYPE_DOUBLE:
			blr[n] = 27
			n++
		case SQL_TYPE_FLOAT:
			blr[n] = 10
			n++
		case SQL_TYPE_D_FLOAT:
			blr[n] = 11
			n++
		case SQL_TYPE_DATE:
			blr[n] = 12
			n++
		case SQL_TYPE_TIME:
			blr[n] = 13
			n++
		case SQL_TYPE_TIMESTAMP:
			blr[n] = 35
			n++
		case SQL_TYPE_BLOB:
			blr[n] = 9
			blr[n+1] = 0
			n += 2
		case SQL_TYPE_ARRAY:
			blr[n] = 9
			blr[n+1] = 0
			n += 2
		case SQL_TYPE_BOOLEAN:
			blr[n] = 23
			n++
		case SQL_TYPE_DEC64:
			blr[n] = 24
			n++
		case SQL_TYPE_DEC128:
			blr[n] = 25
			n++
		case SQL_TYPE_TIME_TZ:
			blr[n] = 28
			n++
		case SQL_TYPE_TIMESTAMP_TZ:
			blr[n] = 29
			n++
		}
		// [blr_short, 0]
		blr[n] = 7
		blr[n+1] = 0
		n += 2
	}
	// [blr_end, blr_eoc]
	blr[n] = 255
	blr[n+1] = 76
	n += 2

	return blr[:n]
}
