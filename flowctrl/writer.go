package flowctrl

import (
	"io"
)

type rateWriter struct {
	underlying io.Writer
	c          *Controller
}

func (w *rateWriter) Write(p []byte) (written int, err error) {
	for {
		size := len(p)
		size = w.c.acquire(size)

		n, writeErr := w.underlying.Write(p[:size])
		w.c.fill(size - n)
		written += n
		if writeErr != nil {
			err = writeErr
			return
		}
		if size == len(p) {
			return
		}
		p = p[size:]
	}
}

func NewRateWriter(w io.Writer, ratePerSecond int) io.WriteCloser {
	c := NewController(ratePerSecond)
	w = c.Writer(w)
	return struct {
		io.Writer
		io.Closer
	}{
		Writer: w,
		Closer: c,
	}
}

func NewRateWriterWithCtrl(w io.Writer, c *Controller) io.Writer {
	w = c.Writer(w)
	return w
}
