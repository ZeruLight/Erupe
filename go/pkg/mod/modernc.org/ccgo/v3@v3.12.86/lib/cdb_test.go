package ccgo

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestCDBWriter(t *testing.T) {
	var b bytes.Buffer

	wr := newCDBWriter(&b)

	items := []cdbItem{
		{
			Arguments: []string{"hello", "there"},
			Directory: "/work",
		},
		{
			Arguments: []string{"good", "bye"},
			Directory: "/work/src",
		},
	}

	for _, it := range items {
		wr.add(it)
	}

	if err := wr.finish(); err != nil {
		t.Fatal(err)
	}

	var got []cdbItem
	if err := json.Unmarshal(b.Bytes(), &got); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, items) {
		t.Errorf("got items\n%#v\nwant\n%#v", got, items)
	}

	if !strings.HasPrefix(b.String(), "[\n  ") {
		t.Errorf("got non-pretty-printed output:\n%s", b.String())
	}
}

func TestMakeDParser(t *testing.T) {
	in := `CreateProcess(C:\Program Files\CodeBlocks\MinGW\bin\gcc.exe,gcc -O3 -Wall -c -o adler32.o adler32.c,...)`
	got, err := makeDParser(in)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{`C:\Program Files\CodeBlocks\MinGW\bin\gcc.exe`, "-O3", "-Wall", "-c", "-o", "adler32.o", "adler32.c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got args %v\nwant %v", got, want)
	}
}

func TestStraceParser(t *testing.T) {
	in := `execve("/usr/bin/ar", ["ar", "cr", "libtcl8.6.a", "regcomp.o", "bn_s_mp_sqr.o", "bn_s_mp_sub.o"], 0x55e6bbf49648 /* 60 vars */) = 0`
	got, err := straceParser(in)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"/usr/bin/ar", "cr", "libtcl8.6.a", "regcomp.o", "bn_s_mp_sqr.o", "bn_s_mp_sub.o"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got args %v\nwant %v", got, want)
	}
}
