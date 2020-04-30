package buffered

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"time"
)

// AsyncBufferWriteSyncer will stores output in a buffer and flushes the buffer either when
// 1) flushInterval is expired, or 2) the accumulated buffer has reached the max bufferSize
type AsyncBufferWriteSyncer struct {
	sync.Mutex
	bufWriter     *bufio.Writer
	w             io.Writer // used to reset the bufWriter when needed
	flushInterval time.Duration
	ticker        time.Ticker
}

// NewAsyncBufferWriteSyncer creates a new AsyncBufferWriteSyncer.  Once the accumulated buffer has past bufferSize,
// the buffer is flushed and timer reset.  The buffer would regularly be flushed in the flushInterval regardless if
// the buffer has past the bufferSize.
func NewAsyncBufferWriteSyncer(w io.Writer, bufferSize int, flushInterval time.Duration) *AsyncBufferWriteSyncer {
	bufWriter := bufio.NewWriterSize(w, bufferSize)

	a := &AsyncBufferWriteSyncer{
		bufWriter:     bufWriter,
		w:             w,
		flushInterval: flushInterval,
		ticker:        *time.NewTicker(flushInterval),
	}

	go a.run()

	return a
}

// Write implements io.Writer.Write()
func (a *AsyncBufferWriteSyncer) Write(p []byte) (int, error) {
	a.Lock()
	// bufWriter.Write will flushe the buffer if len(p) > available buffer
	n, err := a.bufWriter.Write(p)
	if err != nil {
		fmt.Println("level=ERROR, module=asyncBufferWriteSyncer, src=write, err=", err.Error())
		// Reset the Writer because https://golang.org/pkg/bufio/#Writer
		// "If an error occurs writing to a Writer, no more data will be accepted and all subsequent writes, and Flush, will return the error."
		a.bufWriter.Reset(a.w)
	}
	a.Unlock()
	return n, err
}

// Sync implements Syncer and flushes the buffer
func (a *AsyncBufferWriteSyncer) Sync() error {
	a.Lock()
	err := a.bufWriter.Flush()
	if err != nil {
		fmt.Println("level=ERROR, module=asyncBufferWriteSyncer, src=sync, err=", err.Error())
		// Reset the Writer because https://golang.org/pkg/bufio/#Writer
		// "If an error occurs writing to a Writer, no more data will be accepted and all subsequent writes, and Flush, will return the error."
		a.bufWriter.Reset(a.w)
	}
	a.Unlock()
	return err
}

// fun listens to the tick channel and flushes the buffer
func (a *AsyncBufferWriteSyncer) run() {
	for {
		select {
		case <-a.ticker.C:
			if a.bufWriter.Buffered() > 0 {
				a.Sync()
			}
		}
	}
}
