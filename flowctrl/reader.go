package flowctrl

import (
	"io"
)

type rateReader struct {
	underlying io.Reader
	c          *Controller
}

func (r *rateReader) Read(p []byte) (n int, err error) {
	size := len(p)
	size = r.c.acquire(size)
	n, err = r.underlying.Read(p[:size])
	r.c.fill(size - n)
	return
}

func (r *rateReader) Close() error {
	if closer, ok := r.underlying.(io.ReadCloser); ok {
		return closer.Close()
	}
	return nil
}

func NewRateReader(r io.Reader, ratePerSecond int) io.ReadCloser {
	c := NewController(ratePerSecond)
	r = c.Reader(r)
	return struct {
		io.Reader
		io.Closer
	}{
		Reader: r,
		Closer: c,
	}
}

func NewRateReaderWithCtrl(r io.Reader, c *Controller) io.Reader {
	r = c.Reader(r)
	return r
}
