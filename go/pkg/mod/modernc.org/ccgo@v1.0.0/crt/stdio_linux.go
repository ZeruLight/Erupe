// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crt // import "modernc.org/ccgo/crt"

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"sync"
	"unsafe"

	"modernc.org/ccir/libc/errno"
	"modernc.org/ccir/libc/stdio"
	"modernc.org/internal/buffer"
	"modernc.org/mathutil"
)

var (
	stdin, stdout, stderr unsafe.Pointer
)

var (
	files = &fmap{
		m: map[unsafe.Pointer]*os.File{},
	}
	nullReader = bytes.NewBuffer(nil)
)

type fmap struct {
	m  map[unsafe.Pointer]*os.File
	mu sync.Mutex
}

func (m *fmap) add(f *os.File, u unsafe.Pointer) {
	m.mu.Lock()
	m.m[u] = f
	m.mu.Unlock()
}

func (m *fmap) reader(u unsafe.Pointer) io.Reader {
	switch u {
	case stdin:
		return os.Stdin
	case stdout, stderr:
		return nullReader
	}

	m.mu.Lock()
	f := m.m[u]
	m.mu.Unlock()
	return f
}

func (m *fmap) file(u unsafe.Pointer) *os.File {
	switch u {
	case stdin:
		return os.Stdin
	case stdout:
		return os.Stdout
	case stderr:
		return os.Stderr
	}

	m.mu.Lock()
	f := m.m[u]
	m.mu.Unlock()
	return f
}

func (m *fmap) writer(u unsafe.Pointer) io.Writer {
	switch u {
	case stdin:
		return ioutil.Discard
	case stdout:
		return os.Stdout
	case stderr:
		return os.Stderr
	}

	m.mu.Lock()
	f := m.m[u]
	m.mu.Unlock()
	return f
}

func (m *fmap) extract(u unsafe.Pointer) *os.File {
	m.mu.Lock()
	f := m.m[u]
	delete(m.m, u)
	m.mu.Unlock()
	return f
}

// void __register_stdfiles(void *, void *, void *);
func X__register_stdfiles(tls *TLS, in, out, err unsafe.Pointer) {
	stdin = in
	stdout = out
	stderr = err
}

// int printf(const char *format, ...);
func X__builtin_printf(tls *TLS, format *int8, args ...interface{}) int32 {
	return goFprintf(os.Stdout, format, args...)
}

// int printf(const char *format, ...);
func Xprintf(tls *TLS, format *int8, args ...interface{}) int32 {
	return X__builtin_printf(tls, format, args...)
}

// int sprintf(char *str, const char *format, ...);
func X__builtin_sprintf(tls *TLS, str, format *int8, args ...interface{}) int32 {
	w := memWriter(unsafe.Pointer(str))
	n := goFprintf(&w, format, args...)
	w.WriteByte(0)
	return n
}

// int sprintf(char *str, const char *format, ...);
func Xsprintf(tls *TLS, str, format *int8, args ...interface{}) int32 {
	return X__builtin_sprintf(tls, str, format, args...)
}

func goFprintf(w io.Writer, format *int8, ap ...interface{}) int32 {
	var b buffer.Bytes
	written := 0
	for {
		c := *format
		*(*uintptr)(unsafe.Pointer(&format))++
		switch c {
		case 0:
			_, err := b.WriteTo(w)
			b.Close()
			if err != nil {
				return -1
			}

			return int32(written)
		case '%':
			modifiers := ""
			long := 0
			var w []interface{}
		more:
			c := *format
			*(*uintptr)(unsafe.Pointer(&format))++
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
				modifiers += string(c)
				goto more
			case '*':
				w = append(w, VAInt32(&ap))
				modifiers += string(c)
				goto more
			case 'c':
				arg := VAInt32(&ap)
				n, _ := fmt.Fprintf(&b, fmt.Sprintf("%%%sc", modifiers), append(w, arg)...)
				written += n
			case 'd', 'i':
				var arg interface{}
				switch long {
				case 0:
					arg = VAInt32(&ap)
				case 1:
					arg = VALong(&ap)
				default:
					arg = VAInt64(&ap)
				}
				n, _ := fmt.Fprintf(&b, fmt.Sprintf("%%%sd", modifiers), append(w, arg)...)
				written += n
			case 'u':
				var arg interface{}
				switch long {
				case 0:
					arg = VAUint32(&ap)
				case 1:
					arg = VAULong(&ap)
				default:
					arg = VAUint64(&ap)
				}
				n, _ := fmt.Fprintf(&b, fmt.Sprintf("%%%sd", modifiers), append(w, arg)...)
				written += n
			case 'x':
				var arg interface{}
				switch long {
				case 0:
					arg = VAUint32(&ap)
				case 1:
					arg = VAULong(&ap)
				default:
					arg = VAUint64(&ap)
				}
				n, _ := fmt.Fprintf(&b, fmt.Sprintf("%%%sx", modifiers), append(w, arg)...)
				written += n
			case 'l':
				long++
				goto more
			case 'f':
				arg := VAFloat64(&ap)
				n, _ := fmt.Fprintf(&b, fmt.Sprintf("%%%sf", modifiers), append(w, arg)...)
				written += n
			case 'p':
				arg := VAPointer(&ap)
				n, _ := fmt.Fprintf(&b, fmt.Sprintf("%%%sp", modifiers), append(w, arg)...)
				written += n
			case 'g':
				arg := VAFloat64(&ap)
				n, _ := fmt.Fprintf(&b, fmt.Sprintf("%%%sg", modifiers), append(w, arg)...)
				written += n
			case 's':
				arg := (*int8)(VAPointer(&ap))
				if arg == nil {
					break
				}

				var b2 buffer.Bytes
				for {
					c := *arg
					*(*uintptr)(unsafe.Pointer(&arg))++
					if c == 0 {
						n, _ := fmt.Fprintf(&b, fmt.Sprintf("%%%ss", modifiers), append(w, b2.Bytes())...)
						b2.Close()
						written += n
						break
					}

					b2.WriteByte(byte(c))
				}
			default:
				panic(fmt.Errorf("TODO %q", "%"+string(c)))
			}
		default:
			b.WriteByte(byte(c))
			written++
			if c == '\n' {
				if _, err := b.WriteTo(w); err != nil {
					b.Close()
					return -1
				}
				b.Reset()
			}
		}
	}
}

// FILE *fopen64(const char *path, const char *mode);
func Xfopen64(tls *TLS, path, mode *int8) *XFILE {
	p := GoString(path)
	var u unsafe.Pointer
	switch p {
	case os.Stderr.Name():
		u = stderr
	case os.Stdin.Name():
		u = stdin
	case os.Stdout.Name():
		u = stdout
	default:
		var f *os.File
		var err error
		switch mode := GoString(mode); mode {
		case "a":
			if f, err = os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
				switch {
				case os.IsPermission(err):
					tls.setErrno(errno.XEPERM)
				default:
					tls.setErrno(errno.XEACCES)
				}
			}
		case "r", "rb":
			if f, err = os.OpenFile(p, os.O_RDONLY, 0666); err != nil {
				switch {
				case os.IsNotExist(err):
					tls.setErrno(errno.XENOENT)
				case os.IsPermission(err):
					tls.setErrno(errno.XEPERM)
				default:
					tls.setErrno(errno.XEACCES)
				}
			}
		case "w":
			if f, err = os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666); err != nil {
				switch {
				case os.IsPermission(err):
					tls.setErrno(errno.XEPERM)
				default:
					tls.setErrno(errno.XEACCES)
				}
			}
		default:
			panic(mode)
		}
		if f != nil {
			u = malloc(tls, ptrSize)
			files.add(f, u)
		}
	}
	return (*XFILE)(u)
}

// size_t fwrite(const void *ptr, size_t size, size_t nmemb, FILE *stream);
func fwrite(tls *TLS, ptr unsafe.Pointer, size, nmemb uint64, stream *XFILE) uint64 {
	hi, lo := mathutil.MulUint128_64(size, nmemb)
	if hi != 0 || lo > math.MaxInt32 {
		tls.setErrno(errno.XE2BIG)
		return 0
	}

	n, err := files.writer(unsafe.Pointer(stream)).Write((*[math.MaxInt32]byte)(ptr)[:lo])
	if err != nil {
		tls.setErrno(errno.XEIO)
	}
	return uint64(n) / size
}

// int fclose(FILE *stream);
func Xfclose(tls *TLS, stream *XFILE) int32 {
	u := unsafe.Pointer(stream)
	switch u {
	case stdin, stdout, stderr:
		tls.setErrno(errno.XEIO)
		return stdio.XEOF
	}

	f := files.extract(u)
	if f == nil {
		tls.setErrno(errno.XEBADF)
		return stdio.XEOF
	}

	Xfree(tls, u)
	if err := f.Close(); err != nil {
		tls.setErrno(errno.XEIO)
		return stdio.XEOF
	}

	return 0
}

// size_t fread(void *ptr, size_t size, size_t nmemb, FILE *stream);
func fread(tls *TLS, ptr unsafe.Pointer, size, nmemb uint64, stream *XFILE) uint64 {
	hi, lo := mathutil.MulUint128_64(size, nmemb)
	if hi != 0 || lo > math.MaxInt32 {
		tls.setErrno(errno.XE2BIG)
		return 0
	}

	n, err := files.reader(unsafe.Pointer(stream)).Read((*[math.MaxInt32]byte)(ptr)[:lo])
	if err != nil {
		tls.setErrno(errno.XEIO)
	}
	return uint64(n) / size
}

func fseek(tls *TLS, stream *XFILE, offset int64, whence int32) int32 {
	f := files.file(unsafe.Pointer(stream))
	if f == nil {
		tls.setErrno(errno.XEBADF)
		return -1
	}

	if _, err := f.Seek(offset, int(whence)); err != nil {
		tls.setErrno(errno.XEINVAL)
		return -1
	}

	return 0
}

func ftell(tls *TLS, stream *XFILE) int64 {
	f := files.file(unsafe.Pointer(stream))
	if f == nil {
		tls.setErrno(errno.XEBADF)
		return -1
	}

	n, err := f.Seek(0, os.SEEK_CUR)
	if err != nil {
		tls.setErrno(errno.XEBADF)
		return -1
	}

	return n
}

// int fgetc(FILE *stream);
func Xfgetc(tls *TLS, stream *XFILE) int32 {
	p := buffer.Get(1)
	if _, err := files.reader(unsafe.Pointer(stream)).Read(*p); err != nil {
		buffer.Put(p)
		return stdio.XEOF
	}

	r := int32((*p)[0])
	buffer.Put(p)
	return r
}

// char *fgets(char *s, int size, FILE *stream);
func Xfgets(tls *TLS, s *int8, size int32, stream *XFILE) *int8 {
	f := files.reader(unsafe.Pointer(stream))
	p := buffer.Get(1)
	b := *p
	w := memWriter(unsafe.Pointer(s))
	ok := false
	for i := int(size) - 1; i > 0; i-- {
		_, err := f.Read(b)
		if err != nil {
			if !ok {
				buffer.Put(p)
				return nil
			}

			break
		}

		ok = true
		w.WriteByte(b[0])
		if b[0] == '\n' {
			break
		}
	}
	w.WriteByte(0)
	buffer.Put(p)
	return s

}

// int __builtin_fprintf(void* stream, const char *format, ...);
func X__builtin_fprintf(tls *TLS, stream unsafe.Pointer, format *int8, args ...interface{}) int32 {
	return goFprintf(files.writer(stream), format, args...)
}

// int fprintf(FILE * stream, const char *format, ...);
func Xfprintf(tls *TLS, stream *XFILE, format *int8, args ...interface{}) int32 {
	return X__builtin_fprintf(tls, unsafe.Pointer(stream), format, args...)
}

// int fflush(FILE *stream);
func Xfflush(tls *TLS, stream *XFILE) int32 {
	f := files.file(unsafe.Pointer(stream))
	if f == nil {
		tls.setErrno(stdio.XEOF)
		return -1
	}

	if err := f.Sync(); err != nil {
		tls.setErrno(err)
		return -1
	}

	return 0
}

// int vprintf(const char *format, va_list ap);
func Xvprintf(tls *TLS, format *int8, ap []interface{}) int32 {
	return goFprintf(os.Stdout, format, ap...)
}

// int vfprintf(FILE *stream, const char *format, va_list ap);
func Xvfprintf(tls *TLS, stream *XFILE, format *int8, ap []interface{}) int32 {
	return goFprintf(files.writer(unsafe.Pointer(stream)), format, ap...)
}

// void rewind(FILE *stream);
func Xrewind(tls *TLS, stream *XFILE) { fseek(tls, stream, 0, int32(os.SEEK_SET)) }
