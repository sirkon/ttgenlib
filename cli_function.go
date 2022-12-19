package ttgenlib

import "github.com/sirkon/ttgenlib/internal/generator"

// commandFunction command to render function test.
type commandFunction struct {
	Name goIdentifier `arg:"" help:"commandFunction name." predict:"FUNCTION_NAME" required:""`
}

// Run runs command logic.
func (c commandFunction) Run(ctx *runContext) error {
	return generator.GenerateForFunction(
		ctx.args.PkgPath.String(),
		ctx.args.Function.Name.String(),
		ctx.lookup,
		ctx.logging,
		ctx.opts...,
	)
}
