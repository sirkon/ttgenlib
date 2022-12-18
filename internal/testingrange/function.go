package main

import "io"

type theType struct {
	dst io.Writer
}

func (t *theType) someFunction(src io.Reader) ([]byte, error) {
	return io.ReadAll(src)
}
