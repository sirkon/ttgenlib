package main

import (
	"github.com/sirkon/errors"
	"github.com/sirkon/gogh"
	"github.com/sirkon/message"
	"github.com/sirkon/ttgenlib"
)

func main() {
	lr := &loggingRenderer{
		needTestLog: true,
	}
	err := ttgenlib.Run(
		"example",
		ttgenlib.StandardMockLookup([]string{"internal/extmocks"}, "${type|P}Mock", nil),
		lr,
		ttgenlib.GenPreTest(func(r *gogh.GoRenderer[*gogh.Imports]) {
			lr.z = r.Z()
		}),
	)
	if err != nil {
		message.Fatal(errors.Wrap(err, "run command"))
	}
}

type loggingRenderer struct {
	z           *gogh.GoRenderer[*gogh.Imports]
	needTestLog bool
}

func (lr *loggingRenderer) ExpectedError(r *gogh.GoRenderer[*gogh.Imports]) {
	lr.myImports(r)
	r.L(`tl.Log($errs.Wrap(err, "expected error"))`)
}

func (lr *loggingRenderer) UnexpectedError(r *gogh.GoRenderer[*gogh.Imports]) {
	lr.myImports(r)
	r.L(`tl.Error($errs.Wrap(err, "unexpected error"))`)
}

func (lr *loggingRenderer) ErrorWasExpected(r *gogh.GoRenderer[*gogh.Imports]) {
	lr.myImports(r)
	r.L(`tl.Error($errs.New("error was expected"))`)
}

func (lr *loggingRenderer) InvalidError(r *gogh.GoRenderer[*gogh.Imports], errvar string) {
	lr.myImports(r)
	r.L(`tl.Error($errs.Wrap($0, "check error kind"))`, errvar)
}

func (lr *loggingRenderer) myImports(r *gogh.GoRenderer[*gogh.Imports]) {
	if lr.needTestLog {
		lr.needTestLog = false
		r.Imports().Add("github.com/sirkon/testlog").Ref("tl")
		lr.z.L(`tl := $tl.New(t)`)
	}
	r.Imports().Add("github.com/sirkon/errors").Ref("errs")
}
