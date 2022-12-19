package main

import (
	"github.com/sirkon/errors"
	"github.com/sirkon/gogh"
	"github.com/sirkon/message"
	"github.com/sirkon/ttgenlib"
)

func main() {
	err := ttgenlib.Run(
		"example",
		ttgenlib.StandardMockLookup([]string{"internal/extmocks"}, "${type|P}Mock", nil),
		loggingRenderer{},
		ttgenlib.GenPreTest(func(r *gogh.GoRenderer[*gogh.Imports]) {
			r.Imports().Add("github.com/sirkon/testlog").Ref("tl")
			r.L(`tl := testlog.New(t)`)
		}),
	)
	if err != nil {
		message.Fatal(errors.Wrap(err, "run command"))
	}
}

type loggingRenderer struct{}

func (loggingRenderer) ExpectedError(r *gogh.GoRenderer[*gogh.Imports]) {
	myImports(r)
	r.L(`tl.Log($errs.Wrap(err, "expected error"))`)
}

func (loggingRenderer) UnexpectedError(r *gogh.GoRenderer[*gogh.Imports]) {
	myImports(r)
	r.L(`tl.Error($errs.Wrap(err, "unexpected error"))`)
}

func (loggingRenderer) ErrorWasExpected(r *gogh.GoRenderer[*gogh.Imports]) {
	myImports(r)
	r.L(`tl.Error($errs.New("error was expected"))`)
}

func (loggingRenderer) InvalidError(r *gogh.GoRenderer[*gogh.Imports], errvar string) {
	myImports(r)
	r.L(`tl.Error($errs.Wrap($0, "check error kind"))`, errvar)
}

func myImports(r *gogh.GoRenderer[*gogh.Imports]) {
	r.Imports().Add("github.com/sirkon/errors").Ref("errs")
}
