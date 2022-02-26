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

package main

import (
	"github.com/arteev/firebirdsql"
	"log"
	"math/rand"
	"time"
)

func main() {
	dsn := "sysdba:masterkey@127.0.0.1/bar.fdb"
	events := []string{"my_event", "order_created"}
	fbEvent, err := firebirdsql.NewFBEvent(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer fbEvent.Close()

	chEvent := make(chan firebirdsql.Event)
	sbr, err := fbEvent.SubscribeChan(events, chEvent)
	if err != nil {
		log.Fatal(err)
	}
	defer sbr.Unsubscribe()

	go func() {
		for i := 0; i < 100; i++ {
			fbEvent.PostEvent(events[ rand.Intn(2)])
		}
	}()

	for event := range chEvent {
		log.Printf("event: %s, count: %d, id: %d, remote id:%d \n",
			event.Name, event.Count, event.ID, event.RemoteID)
		time.Sleep(time.Millisecond*50) // for example: that the quantity may be greater than one
	}
}
