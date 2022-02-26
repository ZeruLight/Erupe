// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build 386 arm arm64be armbe mips mipsle ppc ppc64le s390 s390x sparc

package crt // import "modernc.org/ccgo/crt"

import (
	"fmt"
	"os"
	"unsafe"
)

const (
	Tpthread_attr_t  = "union{[36]int8,int32}"
	Tpthread_mutex_t = "union{__data struct{__lock int32,__count uint32,__owner int32,__kind int32,__nusers uint32, union{__elision_data struct{__espins int16,__elision int16},__list struct{__next *struct{}}}},__size [24]int8,__align int32}"
)

type Xpthread_mutex_t struct {
	X [0]struct {
		_        *byte
		X__size  [24]int8
		X__align int32
	}
	U [24]byte
} // union{__data struct{__lock int32,__count uint32,__owner int32,__kind int32,__nusers uint32, union{__elision_data struct{__espins int16,__elision int16},__list struct{__next *struct{}}}},__size [24]int8,__align int32}

type Xpthread_attr_t struct {
	X [0]struct {
		X0 [36]int8
		X1 int32
	}
	U [36]byte
}

// pthread_t pthread_self(void);
func Xpthread_self(tls *TLS) uint32 {
	threadID := tls.threadID
	if ptrace {
		fmt.Fprintf(os.Stderr, "pthread_self() %v\n", threadID)
	}
	return uint32(threadID)
}

// extern int pthread_equal(pthread_t __thread1, pthread_t __thread2);
func Xpthread_equal(tls *TLS, thread1, thread2 uint32) int32 {
	if thread1 == thread2 {
		return 1
	}

	var r int32
	if ptrace {
		fmt.Fprintf(os.Stderr, "pthread_equal(%v, %v) %v\n", thread1, thread2, r)
	}
	return r
}

// int pthread_join(pthread_t thread, void **value_ptr);
func Xpthread_join(tls *TLS, thread uint32, value_ptr *unsafe.Pointer) int32 {
	threads.Lock()
	t := threads.m[uintptr(thread)]
	threads.Unlock()
	if t != nil {
		<-t.c
		if value_ptr != nil {
			*value_ptr = t.retval
		}
		threads.Lock()
		delete(threads.m, uintptr(thread))
		threads.Unlock()
	}
	var r int32
	if ptrace {
		fmt.Fprintf(os.Stderr, "pthread_join(%v, %p) %v\n", thread, value_ptr, r)
	}
	return r
}

// int pthread_create(pthread_t *restrict thread, const pthread_attr_t *restrict attr, void *(*start_routine)(void*), void *restrict arg);
func Xpthread_create(tls *TLS, thread *uint32, attr *Xpthread_attr_t, start_routine func(*TLS, unsafe.Pointer) unsafe.Pointer, arg unsafe.Pointer) int32 {
	if attr != nil {
		panic("TODO")
	}

	new := NewTLS()
	*thread = uint32(new.threadID)
	threads.Lock()
	t := &threadState{c: make(chan struct{})}
	threads.m[uintptr(new.threadID)] = t
	threads.Unlock()
	ch := make(chan struct{})
	go func() {
		close(ch)
		t.retval = start_routine(new, arg)
		if ptrace {
			fmt.Fprintf(os.Stderr, "thread #%#x finished: %#p\n", new.threadID, t.retval)
		}
		close(t.c)
		if t.detached {
			threads.Lock()
			delete(threads.m, uintptr(new.threadID))
			threads.Unlock()
			if ptrace {
				fmt.Fprintf(os.Stderr, "thread #%#x was detached", new.threadID)
			}
		}
	}()
	var r int32
	if ptrace {
		fmt.Fprintf(os.Stderr, "pthread_create(%p, %p, fn, %p) %v\n", thread, attr, arg, r)
	}
	<-ch
	return r
}

// int pthread_detach(pthread_t thread);
func Xpthread_detach(tls *TLS, thread uint32) int32 {
	threads.Lock()
	if t := threads.m[uintptr(thread)]; t != nil {
		t.detached = true
	}
	threads.Unlock()
	var r int32
	if ptrace {
		fmt.Fprintf(os.Stderr, "pthread_detach(%v) %v\n", thread, r)
	}
	return r
}
