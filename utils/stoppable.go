package utils

/*
Source of Code: github.com/zet4/catsbutnotreally/utils/stoppable.go

The MIT License (MIT)
Copyright (c) 2017 Aleksandr @zet4 Tihomirov

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

import (
	"net"
	"sync"
)

// StoppableListener provides a TCP Listener that can be stopped
type StoppableListener struct {
	net.Listener
	Stop      chan bool // Any message to this channel will gracefully stop the server
	Stopped   bool      // True if the server was stopped gracefully
	ConnCount counter   // Number of active client connections
}

type watchedConn struct {
	net.Conn
	connCount *counter
}

type counter struct {
	sync.Mutex
	c int
}

// Handle starts accepting connections and waits for stop signal
func Handle(l net.Listener) (sl *StoppableListener) {
	sl = &StoppableListener{Listener: l, Stop: make(chan bool, 1)}

	// Start a goroutine that will wait until the stop channel
	// receives a message then close the Listener to stop accepting
	// new connections (but continue to server the established ones)
	go func() {
		<-sl.Stop
		sl.Stopped = true
		sl.Listener.Close()
	}()
	return
}

// Accept handles a new connection
func (sl *StoppableListener) Accept() (c net.Conn, err error) {
	c, err = sl.Listener.Accept()
	if err != nil {
		return
	}

	// Wrap the returned connection so we're able to observe
	// when it is closed
	c = watchedConn{Conn: c, connCount: &sl.ConnCount}

	// Count it
	sl.ConnCount.Lock()
	sl.ConnCount.c++
	sl.ConnCount.Unlock()
	return
}

func (c *counter) Get() int {
	c.Lock()
	defer c.Unlock()
	return c.c
}

func (w watchedConn) Close() error {
	w.connCount.Lock()
	w.connCount.c--
	w.connCount.Unlock()
	return w.Conn.Close()
}
