// Copyright 2017 The DB Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package db // import "modernc.org/db"

import (
	"testing"

	"modernc.org/file"
)

func dListFill(t testing.TB, db *testDB, in []int) []DList {
	a := make([]DList, len(in))
	for i, v := range in {
		n, err := db.NewDList(8)
		if err != nil {
			t.Fatal(i, n, err)
		}

		if err := db.w8(n.DataOff(), int64(v)); err != nil {
			t.Fatal(err)
		}

		a[i] = n
		if i != 0 {
			if err := n.InsertAfter(a[i-1].Off); err != nil {
				t.Fatal(i, err)
			}
		}
	}
	return a
}

func dListVerify(iTest int, t testing.TB, db *testDB, in []DList, out []int) {
	if len(out) == 0 {
		return
	}

	defer func() {
		for i, v := range in {
			if err := db.Free(v.Off); err != nil {
				t.Error(i)
			}
		}
	}()

	off := in[0].Off
	var prev int64
	for i, ev := range out {
		n, err := db.OpenDList(off)
		if err != nil {
			t.Fatal(iTest, i, err)
		}

		p, err := n.Prev()
		if err != nil {
			t.Fatal(iTest, i, err)
		}

		if g, e := p, prev; g != e {
			t.Fatalf("test #%x, list item %v, got prev %#x, expected %#x", iTest, i, g, e)
		}

		v, err := db.r8(n.DataOff())
		if g, e := v, int64(ev); g != e {
			t.Fatalf("test #%v, list item #%v, got %v, expected %v", iTest, i, g, e)
		}

		prev = off
		if off, err = n.Next(); err != nil {
			t.Fatal(iTest, i, err)
		}

		if off == 0 {
			if i != len(out)-1 {
				t.Fatal(iTest, i)
			}

			break
		}
	}
	if off != 0 {
		t.Fatal(iTest, off)
	}
}

func testDList(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	db, f := tmpDB(t, ts)

	defer f()

	tab := []struct {
		data []int
	}{
		{[]int{10}},
		{[]int{10, 20}},
		{[]int{10, 20, 30}},
		{[]int{10, 20, 30, 40}},
	}

	for iTest, test := range tab {
		in := dListFill(t, db, test.data)
		dListVerify(iTest, t, db, in, test.data)
	}
}

func TestDList(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testDList(t, v.f) }) {
			break
		}
	}
}

func testDListInsertAfter(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	db, f := tmpDB(t, ts)

	defer f()

	tab := []struct {
		in    []int
		index int
		out   []int
	}{
		{[]int{10}, 0, []int{10, -1}},
		{[]int{10, 20}, 0, []int{10, -1, 20}},
		{[]int{10, 20}, 1, []int{10, 20, -1}},
		{[]int{10, 20, 30}, 0, []int{10, -1, 20, 30}},
		{[]int{10, 20, 30}, 1, []int{10, 20, -1, 30}},
		{[]int{10, 20, 30}, 2, []int{10, 20, 30, -1}},
		{[]int{10, 20, 30, 40}, 0, []int{10, -1, 20, 30, 40}},
		{[]int{10, 20, 30, 40}, 1, []int{10, 20, -1, 30, 40}},
		{[]int{10, 20, 30, 40}, 2, []int{10, 20, 30, -1, 40}},
		{[]int{10, 20, 30, 40}, 3, []int{10, 20, 30, 40, -1}},
	}
	for iTest, test := range tab {
		in := dListFill(t, db, test.in)
		i := test.index
		n, err := db.NewDList(8)
		if err != nil {
			t.Fatal(iTest)
		}

		if err := n.w8(n.DataOff(), -1); err != nil {
			t.Fatal(iTest)
		}

		if err := n.InsertAfter(in[i].Off); err != nil {
			t.Fatal(iTest)
		}

		in = append(in[:i+1], append([]DList{n}, in[i+1:]...)...)
		dListVerify(iTest, t, db, in, test.out)
	}
}

func TestDListInsertAfter(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testDListInsertAfter(t, v.f) }) {
			break
		}
	}
}

func testDListInsertBefore(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	db, f := tmpDB(t, ts)

	defer f()

	tab := []struct {
		in    []int
		index int
		out   []int
	}{
		{[]int{10}, 0, []int{-1, 10}},
		{[]int{10, 20}, 0, []int{-1, 10, 20}},
		{[]int{10, 20}, 1, []int{10, -1, 20}},
		{[]int{10, 20, 30}, 0, []int{-1, 10, 20, 30}},
		{[]int{10, 20, 30}, 1, []int{10, -1, 20, 30}},
		{[]int{10, 20, 30}, 2, []int{10, 20, -1, 30}},
		{[]int{10, 20, 30, 40}, 0, []int{-1, 10, 20, 30, 40}},
		{[]int{10, 20, 30, 40}, 1, []int{10, -1, 20, 30, 40}},
		{[]int{10, 20, 30, 40}, 2, []int{10, 20, -1, 30, 40}},
		{[]int{10, 20, 30, 40}, 3, []int{10, 20, 30, -1, 40}},
	}
	for iTest, test := range tab {
		in := dListFill(t, db, test.in)
		i := test.index
		n, err := db.NewDList(8)
		if err != nil {
			t.Fatal(iTest)
		}

		if err := n.w8(n.DataOff(), -1); err != nil {
			t.Fatal(iTest)
		}

		if err := n.InsertBefore(in[i].Off); err != nil {
			t.Fatal(iTest)
		}

		in = append(in[:i], append([]DList{n}, in[i:]...)...)
		dListVerify(iTest, t, db, in, test.out)
	}
}

func TestDListInsertBefore(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testDListInsertBefore(t, v.f) }) {
			break
		}
	}
}

func testDListRemove(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	db, f := tmpDB(t, ts)

	defer f()

	tab := []struct {
		in    []int
		index int
		out   []int
	}{
		{[]int{10}, 0, nil},
		{[]int{10, 20}, 0, []int{20}},
		{[]int{10, 20}, 1, []int{10}},
		{[]int{10, 20, 30}, 0, []int{20, 30}},
		{[]int{10, 20, 30}, 1, []int{10, 30}},
		{[]int{10, 20, 30}, 2, []int{10, 20}},
		{[]int{10, 20, 30, 40}, 0, []int{20, 30, 40}},
		{[]int{10, 20, 30, 40}, 1, []int{10, 30, 40}},
		{[]int{10, 20, 30, 40}, 2, []int{10, 20, 40}},
		{[]int{10, 20, 30, 40}, 3, []int{10, 20, 30}},
	}
	for iTest, test := range tab {
		in := dListFill(t, db, test.in)
		i := test.index
		if err := in[i].Remove(nil); err != nil {
			t.Fatal(iTest)
		}

		in = append(in[:i], in[i+1:]...)
		dListVerify(iTest, t, db, in, test.out)
	}
}

func TestDListRemove(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testDListRemove(t, v.f) }) {
			break
		}
	}
}

func testDListRemoveToEnd(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	db, f := tmpDB(t, ts)

	defer f()

	tab := []struct {
		in    []int
		index int
		out   []int
	}{
		{[]int{10}, 0, nil},
		{[]int{10, 20}, 0, nil},
		{[]int{10, 20}, 1, []int{10}},
		{[]int{10, 20, 30}, 0, nil},
		{[]int{10, 20, 30}, 1, []int{10}},
		{[]int{10, 20, 30}, 2, []int{10, 20}},
		{[]int{10, 20, 30, 40}, 0, nil},
		{[]int{10, 20, 30, 40}, 1, []int{10}},
		{[]int{10, 20, 30, 40}, 2, []int{10, 20}},
		{[]int{10, 20, 30, 40}, 3, []int{10, 20, 30}},
	}
	for iTest, test := range tab {
		in := dListFill(t, db, test.in)
		i := test.index
		if err := in[i].RemoveToLast(nil); err != nil {
			t.Fatal(iTest)
		}

		in = in[:i]
		dListVerify(iTest, t, db, in, test.out)
	}
}

func TestDListRemoveToEnd(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testDListRemoveToEnd(t, v.f) }) {
			break
		}
	}
}

func testDListRemoveToFirst(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	db, f := tmpDB(t, ts)

	defer f()

	tab := []struct {
		in    []int
		index int
		out   []int
	}{
		{[]int{10}, 0, nil},
		{[]int{10, 20}, 0, []int{20}},
		{[]int{10, 20}, 1, nil},
		{[]int{10, 20, 30}, 0, []int{20, 30}},
		{[]int{10, 20, 30}, 1, []int{30}},
		{[]int{10, 20, 30}, 2, nil},
		{[]int{10, 20, 30, 40}, 0, []int{20, 30, 40}},
		{[]int{10, 20, 30, 40}, 1, []int{30, 40}},
		{[]int{10, 20, 30, 40}, 2, []int{40}},
		{[]int{10, 20, 30, 40}, 3, nil},
	}
	for iTest, test := range tab {
		in := dListFill(t, db, test.in)
		i := test.index
		if err := in[i].RemoveToFirst(nil); err != nil {
			t.Fatal(iTest)
		}

		in = in[i+1:]
		dListVerify(iTest, t, db, in, test.out)
	}
}

func TestDListRemoveToFirst(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testDListRemoveToFirst(t, v.f) }) {
			break
		}
	}
}

func benchmarkNewDList(b *testing.B, ts func(t testing.TB) (file.File, func()), dataSize int64) {
	db, f := tmpDB(b, ts)

	defer f()

	a := make([]int64, b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n, err := db.NewDList(dataSize)
		if err != nil {
			b.Fatal(err)
		}

		a[i] = n.Off
	}
	b.StopTimer()
	for _, v := range a {
		if err := db.Free(v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewDList0(b *testing.B) {
	for _, v := range ctors {
		b.Run(v.s, func(b *testing.B) { benchmarkNewDList(b, v.f, 0) })
	}
}

func benchmarkDListInsertAfter(b *testing.B, ts func(t testing.TB) (file.File, func()), dataSize int64) {
	db, f := tmpDB(b, ts)

	defer f()

	r, err := db.NewDList(dataSize)
	if err != nil {
		b.Fatal(err)
	}

	a := make([]DList, b.N)
	for i := range a {
		n, err := db.NewDList(dataSize)
		if err != nil {
			b.Fatal(err)
		}

		a[i] = n
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := a[i].InsertAfter(r.Off); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	if err := r.Free(r.Off); err != nil {
		b.Fatal(err)
	}
	for _, v := range a {
		if err := db.Free(v.Off); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDListInsertAfter(b *testing.B) {
	for _, v := range ctors {
		b.Run(v.s, func(b *testing.B) { benchmarkDListInsertAfter(b, v.f, 0) })
	}
}

func benchmarkDListInsertBefore(b *testing.B, ts func(t testing.TB) (file.File, func()), dataSize int64) {
	db, f := tmpDB(b, ts)

	defer f()

	r, err := db.NewDList(dataSize)
	if err != nil {
		b.Fatal(err)
	}

	a := make([]DList, b.N)
	for i := range a {
		n, err := db.NewDList(dataSize)
		if err != nil {
			b.Fatal(err)
		}

		a[i] = n
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := a[i].InsertBefore(r.Off); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	if err := r.Free(r.Off); err != nil {
		b.Fatal(err)
	}
	for _, v := range a {
		if err := db.Free(v.Off); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDListInsertBefore(b *testing.B) {
	for _, v := range ctors {
		b.Run(v.s, func(b *testing.B) { benchmarkDListInsertBefore(b, v.f, 0) })
	}
}

func benchmarkDListNext(b *testing.B, ts func(t testing.TB) (file.File, func()), dataSize int64) {
	db, f := tmpDB(b, ts)

	defer f()

	n, err := db.NewDList(dataSize)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := n.Next()
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	if err := n.Free(n.Off); err != nil {
		b.Fatal(err)
	}
}

func BenchmarkDListNext(b *testing.B) {
	for _, v := range ctors {
		b.Run(v.s, func(b *testing.B) { benchmarkDListNext(b, v.f, 0) })
	}
}

func benchmarkDListPrev(b *testing.B, ts func(t testing.TB) (file.File, func()), dataSize int64) {
	db, f := tmpDB(b, ts)

	defer f()

	n, err := db.NewDList(dataSize)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := n.Prev()
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	if err := n.Free(n.Off); err != nil {
		b.Fatal(err)
	}
}

func BenchmarkDListPrev(b *testing.B) {
	for _, v := range ctors {
		b.Run(v.s, func(b *testing.B) { benchmarkDListPrev(b, v.f, 0) })
	}
}

func benchmarkDListRemove(b *testing.B, ts func(t testing.TB) (file.File, func()), dataSize int64) {
	db, f := tmpDB(b, ts)

	defer f()

	r, err := db.NewDList(dataSize)
	if err != nil {
		b.Fatal(err)
	}

	a := make([]DList, b.N)
	for i := range a {
		n, err := db.NewDList(dataSize)
		if err != nil {
			b.Fatal(err)
		}

		if err := n.InsertAfter(r.Off); err != nil {
			b.Fatal(err)
		}

		a[i] = n
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := a[i].Remove(nil); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	if err := r.Free(r.Off); err != nil {
		b.Fatal(err)
	}
}

func BenchmarkDListRemove(b *testing.B) {
	for _, v := range ctors {
		b.Run(v.s, func(b *testing.B) { benchmarkDListRemove(b, v.f, 0) })
	}
}
