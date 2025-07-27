package main

import (
	"context"
	"io"
)

type theType struct {
	dst io.Writer
}

func (t *theType) someFunction(ctx context.Context, src io.Reader, count int) ([]byte, error) {
	return io.ReadAll(src)
}
