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
	"sync"
)

type eventManager struct {
	wp         *wireProtocol
	handle     int32
	destructor sync.Once
}

func newEventManager(address string, auxHandle int32) (*eventManager, error) {
	wp, err := newWireProtocol(address, "", "UTF8")
	if err != nil {
		return nil, err
	}
	newManager := &eventManager{
		wp:     wp,
		handle: auxHandle,
	}
	return newManager, nil
}

func (e *eventManager) wait(event *remoteEvent, eventCounts chan<- Event) <-chan error {
	chErr := make(chan error, 1)
	go func() {
		for {
			data, err := e.wp.recvPackets(4)
			if err != nil {
				e.wp.debugPrint("recvPackets:%v:%v", err, data)
				chErr <- err
				return
			}
			op := bytes_to_bint32(data)
			switch op {
			case op_event:
				e.wp.recvPackets(4) //handle
				b, _ := e.wp.recvPackets(4)
				szBuf := bytes_to_bint32(b)
				buffer, _ := e.wp.recvPacketsAlignment(int(szBuf))
				e.wp.recvPackets(8) //ast
				b, _ = e.wp.recvPackets(4)
				eventId := bytes_to_bint32(b)
				e.wp.debugPrint("op_event:%v: event id: %v", buffer, eventId)
				for _, count := range event.getEventCounts(buffer) {
					eventCounts <- count
				}
			default:
				e.wp.debugPrint("unknown operation:%v:%v", op, data)
			}
		}
	}()
	return chErr
}

func (e *eventManager) close() (err error) {
	e.destructor.Do(func() {
		err = e.wp.conn.Close()
	})
	return
}
