package main

import (
	"github.com/sirkon/errors"
	"github.com/sirkon/gogh"
	"github.com/sirkon/message"
	"github.com/sirkon/ttgenlib/internal/generator"
)

func main() {
	err := generator.GenerateForMethod(
		"github.com/sirkon/ttgenlib/internal/testingrange",
		"theType",
		"someFunction",
		generator.StdMockLookup(
			[]string{"internal/testingrange/extmocks"},
			"${type|P}Mock",
			nil,
		),
		generator.WithRenderers(
			func(r *gogh.GoRenderer[*gogh.Imports]) {},
			func(r *gogh.GoRenderer[*gogh.Imports]) {
				r.Imports().Add("context").Ref("ctx")
				r.L(`ctx := $ctx.Background()`)
			},
			customMessages{},
		),
	)
	if err != nil {
		message.Fatal(errors.Wrap(err, "generate table tests for a function"))
	}
}

type customMessages struct {
}

// ExpectedError to satisfy generator.MessagesRenderer
func (customMessages) ExpectedError(r *gogh.GoRenderer[*gogh.Imports]) {
	r.L(`t.Log("expected error:", err)`)
}

// UnexpectedError to satisfy generator.MessagesRenderer
func (customMessages) UnexpectedError(r *gogh.GoRenderer[*gogh.Imports]) {
	r.L(`t.Error("unexpected error:", err)`)
}

// ErrorWasExpected to satisfy generator.MessagesRenderer
func (customMessages) ErrorWasExpected(r *gogh.GoRenderer[*gogh.Imports]) {
	r.L(`t.Error("error was expected")`)
}

// InvalidError to satisfy generator.MessagesRenderer
func (customMessages) InvalidError(r *gogh.GoRenderer[*gogh.Imports], errvar string) {
	r.L(`t.Error("check error:", $0)`, errvar)
}
