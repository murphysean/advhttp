package advhttp

import (
	"io"
)

type ReadCloser struct {
	rc     io.ReadCloser
	length int64
}

func (rc *ReadCloser) Read(p []byte) (int, error) {
	n, err := rc.rc.Read(p)
	rc.length += int64(n)
	return n, err
}

func (rc *ReadCloser) Close() error {
	return rc.rc.Close()
}

func (rc *ReadCloser) Length() int64 {
	return rc.length
}

func NewReadCloser(rc io.ReadCloser) *ReadCloser {
	arc := new(ReadCloser)
	arc.rc = rc
	return arc
}
