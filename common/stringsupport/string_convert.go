package stringsupport

import (
	"bytes"
	"io"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func UTF8ToSJIS(x string) []byte {
	e := japanese.ShiftJIS.NewEncoder()
	xt, _, err := transform.String(e, x)
	if err != nil {
		panic(err)
	}
	return []byte(xt)
}

func SJISToUTF8(b []byte) string {
	d := japanese.ShiftJIS.NewDecoder()
	result, err := io.ReadAll(transform.NewReader(bytes.NewReader(b), d))
	if err != nil {
		panic(err)
	}
	return string(result)
}

func PaddedString(x string, size uint, t bool) []byte {
	if t {
		e := japanese.ShiftJIS.NewEncoder()
		xt, _, err := transform.String(e, x)
		if err != nil {
			return make([]byte, size)
		}
		x = xt
	}
	out := make([]byte, size)
	copy(out, x)
	out[len(out)-1] = 0
	return out
}

func CSVAdd(csv string, v int) string {
	if len(csv) == 0 {
		return strconv.Itoa(v)
	}
	if CSVContains(csv, v) {
		return csv
	} else {
		return csv + "," + strconv.Itoa(v)
	}
}

func CSVRemove(csv string, v int) string {
	s := strings.Split(csv, ",")
	for i, e := range s {
		if e == strconv.Itoa(v) {
			s[i] = s[len(s)-1]
			s = s[:len(s)-1]
		}
	}
	return strings.Join(s, ",")
}

func CSVContains(csv string, v int) bool {
	s := strings.Split(csv, ",")
	for i := 0; i < len(s); i++ {
		j, _ := strconv.ParseInt(s[i], 10, 64)
		if int(j) == v {
			return true
		}
	}
	return false
}

func CSVLength(csv string) int {
	if csv == "" {
		return 0
	}
	s := strings.Split(csv, ",")
	return len(s)
}

func CSVElems(csv string) []int {
	var r []int
	if csv == "" {
		return r
	}
	s := strings.Split(csv, ",")
	for i := 0; i < len(s); i++ {
		j, _ := strconv.ParseInt(s[i], 10, 64)
		r = append(r, int(j))
	}
	return r
}
