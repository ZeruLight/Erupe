// Package lz4 implements compression using lz4.c. This is its test
// suite.
//
// Copyright (c) 2013 CloudFlare, Inc.

package lz4

import (
	"io/ioutil"
	"strings"
	"testing"
	"testing/quick"
)

func TestCompressionHCRatio(t *testing.T) {
	input, err := ioutil.ReadFile("sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	output := make([]byte, CompressBound(input))
	outSize, err := CompressHC(input, output)
	if err != nil {
		t.Fatal(err)
	}

	if want := 4317; want != outSize {
		t.Fatalf("HC Compressed output length != expected: %d != %d", want, outSize)
	}
}

func TestCompressionHCLevels(t *testing.T) {
	input, err := ioutil.ReadFile("sample.txt")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Level   int
		Outsize int
	}{
		{0, 4317},
		{1, 4415},
		{2, 4359},
		{3, 4339},
		{4, 4321},
		{5, 4317},
		{6, 4317},
		{7, 4317},
		{8, 4317},
		{9, 4317},
		{10, 4317},
		{11, 4317},
		{12, 4317},
		{13, 4317},
		{14, 4317},
		{15, 4317},
		{16, 4317},
	}

	for _, tt := range cases {
		output := make([]byte, CompressBound(input))
		outSize, err := CompressHCLevel(input, output, tt.Level)
		if err != nil {
			t.Fatal(err)
		}

		if want := tt.Outsize; want != outSize {
			t.Errorf("HC level %d length != expected: %d != %d",
				tt.Level, want, outSize)
		}
	}
}

func TestCompressionHC(t *testing.T) {
	input := []byte(strings.Repeat("Hello world, this is quite something", 10))
	output := make([]byte, CompressBound(input))
	outSize, err := CompressHC(input, output)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}
	if outSize == 0 {
		t.Fatal("Output buffer is empty.")
	}
	output = output[:outSize]
	decompressed := make([]byte, len(input))
	err = Uncompress(output, decompressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}
	if string(decompressed) != string(input) {
		t.Fatalf("Decompressed output != input: %q != %q", decompressed, input)
	}
}

func TestEmptyCompressionHC(t *testing.T) {
	input := []byte("")
	output := make([]byte, CompressBound(input))

	outSize, err := CompressHC(input, output)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}
	if outSize == 0 {
		t.Fatal("Output buffer is empty.")
	}
	output = output[:outSize]
	decompressed := make([]byte, len(input))
	err = Uncompress(output, decompressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}
	if string(decompressed) != string(input) {
		t.Fatalf("Decompressed output != input: %q != %q", decompressed, input)
	}
}

func TestNoCompressionHC(t *testing.T) {
	input := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	output := make([]byte, CompressBound(input))
	outSize, err := CompressHC(input, output)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}
	if outSize == 0 {
		t.Fatal("Output buffer is empty.")
	}
	output = output[:outSize]
	decompressed := make([]byte, len(input))
	err = Uncompress(output, decompressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}
	if string(decompressed) != string(input) {
		t.Fatalf("Decompressed output != input: %q != %q", decompressed, input)
	}
}

func TestCompressionErrorHC(t *testing.T) {
	input := []byte(strings.Repeat("Hello world, this is quite something", 10))
	output := make([]byte, 0)
	outSize, err := CompressHC(input, output)

	if outSize != 0 {
		t.Fatalf("%d", outSize)
	}

	if err == nil {
		t.Fatalf("Compression should have failed but didn't")
	}

	output = make([]byte, 1)
	_, err = CompressHC(input, output)
	if err == nil {
		t.Fatalf("Compression should have failed but didn't")
	}
}

func TestDecompressionErrorHC(t *testing.T) {
	input := []byte(strings.Repeat("Hello world, this is quite something", 10))
	output := make([]byte, CompressBound(input))
	outSize, err := CompressHC(input, output)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}
	if outSize == 0 {
		t.Fatal("Output buffer is empty.")
	}
	output = output[:outSize]
	decompressed := make([]byte, len(input)-1)
	err = Uncompress(output, decompressed)
	if err == nil {
		t.Fatalf("Decompression should have failed")
	}

	decompressed = make([]byte, 1)
	err = Uncompress(output, decompressed)
	if err == nil {
		t.Fatalf("Decompression should have failed")
	}

	decompressed = make([]byte, 0)
	err = Uncompress(output, decompressed)
	if err == nil {
		t.Fatalf("Decompression should have failed")
	}
}

func TestFuzzHC(t *testing.T) {
	f := func(input []byte) bool {
		output := make([]byte, CompressBound(input))
		outSize, err := CompressHC(input, output)
		if err != nil {
			t.Fatalf("Compression failed: %v", err)
		}
		if outSize == 0 {
			t.Fatal("Output buffer is empty.")
		}
		output = output[:outSize]
		decompressed := make([]byte, len(input))
		err = Uncompress(output, decompressed)
		if err != nil {
			t.Fatalf("Decompression failed: %v", err)
		}
		if string(decompressed) != string(input) {
			t.Fatalf("Decompressed output != input: %q != %q", decompressed, input)
		}

		return true
	}

	conf := &quick.Config{MaxCount: 20000}
	if testing.Short() {
		conf.MaxCount = 1000
	}
	if err := quick.Check(f, conf); err != nil {
		t.Fatal(err)
	}
}
