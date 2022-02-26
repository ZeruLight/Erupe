/*******************************************************************************
The MIT License (MIT)

Copyright (c) 2013-2019 Hajime Nakagami

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
	"container/list"
	"encoding/binary"
	"math/big"
	"strconv"
	"time"
)

func str_to_bytes(s string) []byte {
	return []byte(s)
}

func int32_to_bytes(i32 int32) []byte {
	bs := []byte{
		byte(i32 & 0xFF),
		byte(i32 >> 8 & 0xFF),
		byte(i32 >> 16 & 0xFF),
		byte(i32 >> 24 & 0xFF),
	}
	return bs
}

func bint64_to_bytes(i64 int64) []byte {
	bs := []byte{
		byte(i64 >> 56 & 0xFF),
		byte(i64 >> 48 & 0xFF),
		byte(i64 >> 40 & 0xFF),
		byte(i64 >> 32 & 0xFF),
		byte(i64 >> 24 & 0xFF),
		byte(i64 >> 16 & 0xFF),
		byte(i64 >> 8 & 0xFF),
		byte(i64 & 0xFF),
	}
	return bs
}

func bint32_to_bytes(i32 int32) []byte {
	bs := []byte{
		byte(i32 >> 24 & 0xFF),
		byte(i32 >> 16 & 0xFF),
		byte(i32 >> 8 & 0xFF),
		byte(i32 & 0xFF),
	}
	return bs
}

func int16_to_bytes(i16 int16) []byte {
	bs := []byte{
		byte(i16 & 0xFF),
		byte(i16 >> 8 & 0xFF),
	}
	return bs
}
func bytes_to_str(b []byte) string {
	return string(b)
}

func bytes_to_bint32(b []byte) int32 {
	return int32(binary.BigEndian.Uint32(b))
}

func bytes_to_int32(b []byte) int32 {
	return int32(binary.LittleEndian.Uint32(b))
}

func bytes_to_bint16(b []byte) int16 {
	return int16(binary.BigEndian.Uint16(b))
}

func bytes_to_buint16(b []byte) uint16 {
	return uint16(binary.BigEndian.Uint16(b))
}

func bytes_to_int16(b []byte) int16 {
	return int16(binary.LittleEndian.Uint16(b))
}

func bytes_to_bint64(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

func bytes_to_int64(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}

func bigIntFromHexString(s string) *big.Int {
	ret := new(big.Int)
	ret.SetString(s, 16)
	return ret
}

func bigIntFromString(s string) *big.Int {
	ret := new(big.Int)
	ret.SetString(s, 10)
	return ret
}

func flattenBytes(l *list.List) []byte {
	n := 0
	for e := l.Front(); e != nil; e = e.Next() {
		n += len((e.Value).([]byte))
	}

	bs := make([]byte, n)

	n = 0
	for e := l.Front(); e != nil; e = e.Next() {
		for i, b := range (e.Value).([]byte) {
			bs[n+i] = b
		}
		n += len((e.Value).([]byte))
	}

	return bs
}

func xdrBytes(bs []byte) []byte {
	// XDR encoding bytes
	n := len(bs)
	padding := 0
	if n%4 != 0 {
		padding = 4 - n%4
	}
	buf := make([]byte, 4+n+padding)
	buf[0] = byte(n >> 24 & 0xFF)
	buf[1] = byte(n >> 16 & 0xFF)
	buf[2] = byte(n >> 8 & 0xFF)
	buf[3] = byte(n & 0xFF)
	for i, b := range bs {
		buf[4+i] = b
	}
	return buf
}

func xdrString(s string) []byte {
	// XDR encoding string
	bs := bytes.NewBufferString(s).Bytes()
	return xdrBytes(bs)
}

func _int64ToBlr(i64 int64) ([]byte, []byte) {
	v := bint64_to_bytes(i64)
	blr := []byte{16, 0}

	return blr, v
}

func _int32ToBlr(i32 int32) ([]byte, []byte) {
	v := bint32_to_bytes(i32)
	blr := []byte{8, 0}

	return blr, v
}

func _float64ToBlr(v float64) ([]byte, []byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, v)
	blr := []byte{27}
	return blr, buf.Bytes()
}

func _bytesToBlr(v []byte) ([]byte, []byte) {
	nbytes := len(v)
	pad_length := ((4 - nbytes) & 3)
	padding := make([]byte, pad_length)
	v = bytes.Join([][]byte{
		v,
		padding,
	}, nil)
	blr := []byte{14, byte(nbytes & 255), byte(nbytes >> 8)}
	return blr, v
}

func _convert_date(t time.Time) []byte {
	i := int(t.Month()) + 9
	jy := t.Year() + (i / 12) - 1
	jm := i % 12
	c := jy / 100
	jy -= 100 * c
	j := (146097*c)/4 + (1461*jy)/4 + (153*jm+2)/5 + t.Day() - 678882
	return bint32_to_bytes(int32(j))
}

func _convert_time(t time.Time) []byte {
	v := (t.Hour()*3600+t.Minute()*60+t.Second())*10000 + t.Nanosecond()/100000
	return bint32_to_bytes(int32(v))
}

func _dateToBlr(t time.Time) ([]byte, []byte) {
	v := _convert_date(t)
	blr := []byte{12}
	return blr, v
}

func _timeToBlr(t time.Time) ([]byte, []byte) {
	v := _convert_time(t)
	blr := []byte{13}
	return blr, v
}

func _timestampToBlr(t time.Time) ([]byte, []byte) {
	v := bytes.Join([][]byte{
		_convert_date(t),
		_convert_time(t),
	}, nil)

	blr := []byte{35}
	return blr, v
}

func convertToBool(s string, defaultValue bool) bool {
	v, err := strconv.ParseBool(s)
	if err != nil {
		v = defaultValue
	}
	return v
}
