// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32 arm64 mips64 mips64le mips64p32 mips64p32le ppc64 sparc64
// +build !windows

package crt // import "modernc.org/ccgo/crt"

import (
	"unsafe"
)

const (
	TFILE = "struct{_flags int32,_IO_read_ptr *int8,_IO_read_end *int8,_IO_read_base *int8,_IO_write_base *int8,_IO_write_ptr *int8,_IO_write_end *int8,_IO_buf_base *int8,_IO_buf_end *int8,_IO_save_base *int8,_IO_backup_base *int8,_IO_save_end *int8,_markers *struct{},_chain *struct{},_fileno int32,_flags2 int32,_old_offset int64,_cur_column uint16,_vtable_offset int8,_shortbuf [1]int8,_lock *struct{},_offset int64,__pad1 *struct{},__pad2 *struct{},__pad3 *struct{},__pad4 *struct{},__pad5 uint64,_mode int32,_unused2 [20]int8}"
)

type XFILE struct {
	X_flags          int32
	X_IO_read_ptr    *int8
	X_IO_read_end    *int8
	X_IO_read_base   *int8
	X_IO_write_base  *int8
	X_IO_write_ptr   *int8
	X_IO_write_end   *int8
	X_IO_buf_base    *int8
	X_IO_buf_end     *int8
	X_IO_save_base   *int8
	X_IO_backup_base *int8
	X_IO_save_end    *int8
	X_markers        unsafe.Pointer
	X_chain          unsafe.Pointer
	X_fileno         int32
	X_flags2         int32
	X_old_offset     int64
	X_cur_column     uint16
	X_vtable_offset  int8
	X_shortbuf       [1]int8
	X_lock           unsafe.Pointer
	X_offset         int64
	X__pad1          unsafe.Pointer
	X__pad2          unsafe.Pointer
	X__pad3          unsafe.Pointer
	X__pad4          unsafe.Pointer
	X__pad5          uint64
	X_mode           int32
	X_unused2        [20]int8
}

// size_t fwrite(const void *ptr, size_t size, size_t nmemb, FILE *stream);
func Xfwrite(tls *TLS, ptr unsafe.Pointer, size, nmemb uint64, stream *XFILE) uint64 {
	return fwrite(tls, ptr, size, nmemb, stream)
}

// size_t fread(void *ptr, size_t size, size_t nmemb, FILE *stream);
func Xfread(tls *TLS, ptr unsafe.Pointer, size, nmemb uint64, stream *XFILE) uint64 {
	return fread(tls, ptr, size, nmemb, stream)
}

// int fseek(FILE *stream, long offset, int whence);
func Xfseek(tls *TLS, stream *XFILE, offset int64, whence int32) int32 {
	return fseek(tls, stream, offset, whence)
}

// long ftell(FILE *stream);
func Xftell(tls *TLS, stream *XFILE) int64 { return ftell(tls, stream) }
