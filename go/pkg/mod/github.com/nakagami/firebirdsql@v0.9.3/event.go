// +build !plan9

/*******************************************************************************
The MIT License (MIT)

Copyright (c) 2019 Arteev Aleksey

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
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// Errors
var (
	ErrAlreadySubscribe = errors.New("already subscribe")
	ErrFbEventClosed    = errors.New("fbevent already closed")
)

//SQLs
const (
	sqlPostEvent = `execute block as begin post_event '%s'; end`
)

// FbEvent allows you to subscribe to events, also stores subscribers.
// It is possible to send events to the database.
type FbEvent struct {
	mu               sync.RWMutex
	dsn              *firebirdDsn
	conn             *sql.DB
	done             chan struct{}
	closed           int32
	closer           sync.Once
	chDoneSubscriber chan *Subscription
	subscribers      []*Subscription
}

// Event stores event data: the amount since the last time the event was received and id
type Event struct {
	Name     string
	Count    int
	ID       int32
	RemoteID int32
}

// EventHandler callback function type
type EventHandler func(e Event)

// NewFBEvent returns FbEvent for event subscription
func NewFBEvent(dsns string) (*FbEvent, error) {
	conn, err := sql.Open("firebirdsql", dsns)
	if err != nil {
		return nil, err
	}
	// can ignore error, would have been thrown by sql.Open
	dsn, _ := parseDSN(dsns)
	fbEvent := &FbEvent{
		dsn:              dsn,
		conn:             conn,
		done:             make(chan struct{}),
		chDoneSubscriber: make(chan *Subscription),
	}
	go fbEvent.run()
	return fbEvent, nil
}

// PostEvent posts an event to the database
func (e *FbEvent) PostEvent(name string) error {
	_, err := e.conn.Exec(fmt.Sprintf(sqlPostEvent, name))
	if err != nil {
		return err
	}
	return nil
}

func (e *FbEvent) newSubscriber(events []string, cb EventHandler, chEvent chan Event) (*Subscription, error) {
	subscriber, err := newSubscription(e.dsn, events, cb, chEvent, e.chDoneSubscriber)
	if err != nil {
		return nil, err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.subscribers = append(e.subscribers, subscriber)
	return subscriber, nil
}

// Subscribers returns slice of all subscribers
func (e *FbEvent) Subscribers() []*Subscription {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.subscribers[:]
}

// Count returns the number of subscribers
func (e *FbEvent) Count() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.subscribers)
}

// Subscribe subscribe to events using the callback function
func (e *FbEvent) Subscribe(events []string, cb EventHandler) (*Subscription, error) {
	return e.newSubscriber(events, cb, nil)
}

// SubscribeChan subscribe to events using the channel
func (e *FbEvent) SubscribeChan(events []string, chEvent chan Event) (*Subscription, error) {
	return e.newSubscriber(events, nil, chEvent)
}

func (e *FbEvent) run() {
	for {
		select {
		case <-e.done:
			return
		case subscriber := <-e.chDoneSubscriber:
			e.shutdownSubscriber(subscriber)
		}
	}
}

func (e *FbEvent) shutdownSubscriber(subscriber *Subscription) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i := range e.subscribers {
		if e.subscribers[i] == subscriber {
			last := len(e.subscribers) - 1
			e.subscribers[i] = e.subscribers[last]
			e.subscribers[last] = nil
			e.subscribers = e.subscribers[:last]
			return
		}
	}
}

// IsClosed returns a close flag
func (e *FbEvent) IsClosed() bool {
	return atomic.LoadInt32(&e.closed) == 1
}

// Close closes FbEvent and all subscribers
func (e *FbEvent) Close() error {
	if e.IsClosed() {
		return ErrFbEventClosed
	}
	return e.doClose(nil)
}

func (e *FbEvent) closeWithError(err error) error {
	if e.IsClosed() {
		return ErrFbEventClosed
	}
	return e.doClose(err)
}

func (e *FbEvent) doClose(err error) (errResult error) {
	atomic.StoreInt32(&e.closed, 1)
	e.closer.Do(func() {
		e.conn.Close()
		e.mu.Lock()
		wg := &sync.WaitGroup{}
		wg.Add(len(e.subscribers))
		for i := range e.subscribers {
			go func(subscriber *Subscription) {
				defer wg.Done()
				subscriber.unsubscribeNoNotify()
			}(e.subscribers[i])
		}
		e.subscribers = make([]*Subscription, 0)
		e.mu.Unlock()
		wg.Wait()
		close(e.done)
	})
	return
}
