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
	"bytes"
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrEventAlreadyRunning = errors.New("events are already running")
	ErrEventNeed           = errors.New("at least one event is needed")
	ErrWrongLengthEvent    = errors.New("length name events are longer than 255")
	ErrEventBufferLarge    = errors.New("whole events buffer is bigger than 65535")
)

const (
	maxEpbLength       = 65535
	maxEventNameLength = 255
)

type remoteEvent struct {
	mu         sync.RWMutex
	id         int32
	rid        int32
	events     []string
	counts     map[string]int
	prevCounts map[string]int
	running    int32
}

func newRemoteEvent() *remoteEvent {
	newEvent := &remoteEvent{
		events: []string{},
	}
	return newEvent
}

func (e *remoteEvent) queueEvents(events ...string) error {
	if atomic.LoadInt32(&e.running) == 1 {
		return ErrEventAlreadyRunning
	}

	if len(events) == 0 {
		return ErrEventNeed
	}

	for _, event := range events {
		if len(event) > maxEventNameLength {
			return ErrWrongLengthEvent
		}
	}

	if len(buildEpbSlice(events, map[string]int{})) > maxEpbLength {
		return ErrEventBufferLarge
	}

	atomic.StoreInt32(&e.running, 1)
	e.counts = make(map[string]int, len(events))
	e.prevCounts = make(map[string]int, len(events))
	for _, event := range events {
		e.events = append(e.events, event)
		e.counts[event] = 0
		e.prevCounts[event] = 0
	}
	return nil
}

func (e *remoteEvent) cancelEvents() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.events = make([]string, 0)
	e.counts = nil
	e.prevCounts = nil
	atomic.StoreInt32(&e.running, 0)
}

func (e *remoteEvent) getEventCounts(data []byte) []Event {
	e.mu.Lock()
	e.prevCounts = make(map[string]int, len(e.events))
	for k, v := range e.counts {
		e.prevCounts[k] = v
		e.counts[k] = 0
	}

	for i := 1; i < len(data); {
		length := int(data[i])
		i++
		eventName := string(data[i : i+length])
		i += length
		_, found := e.counts[eventName]
		if found {
			e.counts[eventName] = int(bytes_to_int32(data[i:]) - 1)
		}
		i += 4
	}
	e.mu.Unlock()

	e.mu.RLock()
	defer e.mu.RUnlock()
	var result []Event
	for i := 0; i < len(e.events); i++ {
		count := e.counts[e.events[i]] - e.prevCounts[e.events[i]]
		if count == 0 {
			continue
		}
		result = append(result, Event{
			Name:     e.events[i],
			Count:    count,
			ID:       atomic.LoadInt32(&e.id),
			RemoteID: atomic.LoadInt32(&e.rid),
		})
	}
	return result
}

func buildEpbSlice(events []string, counts map[string]int) []byte {
	buf := &bytes.Buffer{}
	buf.WriteByte(byte(EPB_version1))
	for _, event := range events {
		buf.WriteByte(byte(len(event)))
		buf.WriteString(event)
		buf.Write(int32_to_bytes(int32(counts[event] + 1)))
	}
	return buf.Bytes()
}

func (e *remoteEvent) buildEpb() []byte {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return buildEpbSlice(e.events, e.counts)
}
