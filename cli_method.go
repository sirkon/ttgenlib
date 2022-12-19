package ttgenlib

import "github.com/sirkon/ttgenlib/internal/generator"

// commandMethod command to render type's method test.
type commandMethod struct {
	Type goIdentifier `arg:"" help:"Type name." predict:"TYPE_NAME" required:""`
	Name goIdentifier `arg:"" help:"Method name." predict:"METHOD_NAME" required:""`
}

// Run runs command logic.
func (c commandMethod) Run(ctx *runContext) error {
	return generator.GenerateForMethod(
		ctx.args.PkgPath.String(),
		ctx.args.Method.Type.String(),
		ctx.args.Method.Name.String(),
		ctx.lookup,
		ctx.logging,
		ctx.opts...,
	)
}
