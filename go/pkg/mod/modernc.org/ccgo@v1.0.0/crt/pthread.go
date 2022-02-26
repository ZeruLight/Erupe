// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crt // import "modernc.org/ccgo/crt"

import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"modernc.org/ccir/libc/errno"
	"modernc.org/ccir/libc/pthread"
)

const (
	Tpthread_cond_t      = "union{__data struct{__lock int32,__futex uint32,__total_seq uint64,__wakeup_seq uint64,__woken_seq uint64,__mutex *struct{},__nwaiters uint32,__broadcast_seq uint32},__size [48]int8,__align int64}"
	Tpthread_mutexattr_t = "union{__size [4]int8,__align int32}"
)

type Xpthread_cond_t struct {
	X [0]struct {
		_        *byte
		X__size  [48]int8
		X__align int64
	}
	U [48]byte
} // t4 union{__data struct{__lock int32,__futex uint32,__total_seq uint64,__wakeup_seq uint64,__woken_seq uint64,__mutex *struct{},__nwaiters uint32,__broadcast_seq uint32},__size [48]int8,__align int64}

type Xpthread_mutexattr_t struct {
	X [0]struct {
		X__size  [4]int8
		X__align int32
	}
	U [4]byte
} // t175 union{__size [4]int8,__align int32}

type mu struct {
	*sync.Cond
	attr  int32
	count int
	owner uintptr
	sync.Mutex
}

type mutexMap struct {
	m map[unsafe.Pointer]*mu
	sync.Mutex
}

func (m *mutexMap) mu(p unsafe.Pointer) *mu {
	m.Lock()
	r := m.m[p]
	if r == nil {
		r = &mu{}
		r.Cond = sync.NewCond(&r.Mutex)
		m.m[p] = r
	}
	m.Unlock()
	return r
}

type condMap struct {
	m map[unsafe.Pointer]*sync.Cond
	sync.Mutex
}

func (m *condMap) cond(p unsafe.Pointer, mu *mu) *sync.Cond {
	m.Lock()
	r := m.m[p]
	if r == nil {
		r = sync.NewCond(&mu.Mutex)
		m.m[p] = r
	}
	m.Unlock()
	return r
}

type threadState struct {
	c        chan struct{}
	detached bool
	retval   unsafe.Pointer
}

type threadMap struct {
	m map[uintptr]*threadState
	sync.Mutex
}

var (
	conds   = &condMap{m: map[unsafe.Pointer]*sync.Cond{}}
	mutexes = &mutexMap{m: map[unsafe.Pointer]*mu{}}
	threads = &threadMap{m: map[uintptr]*threadState{}}
)

// extern int pthread_mutexattr_init(pthread_mutexattr_t * __attr);
func Xpthread_mutexattr_init(tls *TLS, attr *Xpthread_mutexattr_t) int32 {
	var r int32
	if ptrace {
		fmt.Fprintf(os.Stderr, "pthread_mutexattr_init(%#x) %v\n", attr, r)
	}
	return r
}

// extern int pthread_mutexattr_settype(pthread_mutexattr_t * __attr, int __kind);
func Xpthread_mutexattr_settype(tls *TLS, attr *Xpthread_mutexattr_t, kind int32) int32 {
	*(*int32)(unsafe.Pointer(attr)) = kind
	var r int32
	if ptrace {
		fmt.Fprintf(os.Stderr, "pthread_mutexattr_settype(%#x, %v) %v\n", attr, kind, r)
	}
	return r
}

// extern int pthread_mutex_init(pthread_mutex_t * __mutex, pthread_mutexattr_t * __mutexattr);
func Xpthread_mutex_init(tls *TLS, mutex *Xpthread_mutex_t, mutexattr *Xpthread_mutexattr_t) int32 {
	attr := int32(pthread.XPTHREAD_MUTEX_NORMAL)
	if mutexattr != nil {
		attr = *(*int32)(unsafe.Pointer(mutexattr))
	}
	mutexes.mu(unsafe.Pointer(mutex)).attr = attr
	var r int32
	if ptrace {
		fmt.Fprintf(os.Stderr, "pthread_mutex_init(%p, %#x) %v\n", mutex, mutexattr, r)
	}
	return r
}

// extern int pthread_mutexattr_destroy(pthread_mutexattr_t * __attr);
func Xpthread_mutexattr_destroy(tls *TLS, attr *Xpthread_mutexattr_t) int32 {
	*(*int32)(unsafe.Pointer(attr)) = -1
	var r int32
	if ptrace {
		fmt.Fprintf(os.Stderr, "pthread_mutexattr_destroy(%#x) %v\n", attr, r)
	}
	return r
}

// extern int pthread_mutex_destroy(pthread_mutex_t * __mutex);
func Xpthread_mutex_destroy(tls *TLS, mutex *Xpthread_mutex_t) int32 {
	mutexes.Lock()
	delete(mutexes.m, unsafe.Pointer(mutex))
	mutexes.Unlock()
	var r int32
	if ptrace {
		fmt.Fprintf(os.Stderr, "pthread_mutex_destroy(%p) %v\n", mutex, r)
	}
	return r
}

// extern int pthread_mutex_lock(pthread_mutex_t * __mutex);
func Xpthread_mutex_lock(tls *TLS, mutex *Xpthread_mutex_t) int32 {
	threadID := tls.threadID
	mu := mutexes.mu(unsafe.Pointer(mutex))
	var r int32
	mu.Lock()
	switch mu.attr {
	case pthread.XPTHREAD_MUTEX_NORMAL:
		if mu.count == 0 {
			mu.owner = threadID
			mu.count = 1
			break
		}

		for mu.count != 0 {
			mu.Cond.Wait()
		}
		mu.owner = threadID
		mu.count = 1
	case pthread.XPTHREAD_MUTEX_RECURSIVE:
		if mu.count == 0 {
			mu.owner = threadID
			mu.count = 1
			break
		}

		if mu.owner == threadID {
			mu.count++
			break
		}

		panic("TODO")
	default:
		panic(fmt.Errorf("attr %#x", mu.attr))
	}
	if ptrace {
		fmt.Fprintf(os.Stderr, "pthread_mutex_lock(%p: %+v [thread id %v]) %v\n", mutex, mu, threadID, r)
	}
	mu.Unlock()
	return r
}

// int pthread_mutex_trylock(pthread_mutex_t *mutex);
func Xpthread_mutex_trylock(tls *TLS, mutex *Xpthread_mutex_t) int32 {
	threadID := tls.threadID
	mu := mutexes.mu(unsafe.Pointer(mutex))
	var r int32
	mu.Lock()
	switch mu.attr {
	case pthread.XPTHREAD_MUTEX_NORMAL:
		if mu.count == 0 {
			mu.count = 1
			mu.owner = threadID
			break
		}

		r = errno.XEBUSY
	default:
		panic(fmt.Errorf("attr %#x", mu.attr))
	}
	if ptrace {
		fmt.Fprintf(os.Stderr, "pthread_mutex_trylock(%p: %+v [thread id %v]) %v\n", mutex, mu, threadID, r)
	}
	mu.Unlock()
	return r
}

// extern int pthread_mutex_unlock(pthread_mutex_t * __mutex);
func Xpthread_mutex_unlock(tls *TLS, mutex *Xpthread_mutex_t) int32 {
	threadID := tls.threadID
	mu := mutexes.mu(unsafe.Pointer(mutex))
	var r int32
	mu.Lock()
	switch mu.attr {
	case pthread.XPTHREAD_MUTEX_NORMAL:
		if mu.count == 0 {
			panic("TODO")
		}

		mu.owner = 0
		mu.count = 0
		mu.Cond.Broadcast()
	case pthread.XPTHREAD_MUTEX_RECURSIVE:
		if mu.count == 0 {
			panic("TODO")
		}

		if mu.owner == threadID {
			mu.count--
			if mu.count != 0 {
				break
			}

			mu.owner = 0
			mu.Cond.Broadcast()
			break
		}

		panic("TODO")
	default:
		panic(fmt.Errorf("TODO %#x", mu.attr))
	}
	if ptrace {
		fmt.Fprintf(os.Stderr, "pthread_mutex_unlock(%p: %+v [thread id %v]) %v\n", mutex, mu, threadID, r)
	}
	mu.Unlock()
	return r
}

// int pthread_cond_wait(pthread_cond_t *cond, pthread_mutex_t *mutex);
func Xpthread_cond_wait(tls *TLS, cond *Xpthread_cond_t, mutex *Xpthread_mutex_t) int32 {
	threadID := tls.threadID
	mu := mutexes.mu(unsafe.Pointer(mutex))
	mu.Lock()
	if mu.count == 0 {
		panic("TODO")
	}

	if mu.owner != threadID {
		panic("TODO")
	}

	count := mu.count
	mu.count = 0
	conds.cond(unsafe.Pointer(cond), mu).Wait()
	mu.count = count
	var r int32
	if ptrace {
		fmt.Fprintf(os.Stderr, "pthread_cond_wait(%p, %p) %v\n", cond, mutex, r)
	}
	mu.Unlock()
	return r
}

// int pthread_cond_signal(pthread_cond_t *cond);
func Xpthread_cond_signal(tls *TLS, cond *Xpthread_cond_t) int32 {
	conds.cond(unsafe.Pointer(cond), nil).Signal()
	var r int32
	if ptrace {
		fmt.Fprintf(os.Stderr, "pthread_cond_signal(%p) %v\n", cond, r)
	}
	return r
}
