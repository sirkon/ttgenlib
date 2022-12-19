package ttgenlib

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/posener/complete"
	"github.com/sirkon/ttgenlib/internal/predictor"
	"github.com/willabides/kongplete"
)

// Run table tests generation with a given names and handlers.
func Run(
	appName string,
	mockLookup MockLookup,
	logging GenLoggingRenderer,
	genOpts ...GenOption,
) error {
	globalVarAppName = appName
	var cli cliArgs

	parser := kong.Must(
		&cli,
		kong.Description("Generate table test for a given method of a structure."),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary: true,
			Compact: true,
		}),
		kong.UsageOnError(),
	)
	kongplete.Complete(
		parser,
		kongplete.WithPredictors(map[string]complete.Predictor{
			"PKG_PATH":      predictor.PackagePath{},
			"TYPE_NAME":     predictor.TypeName{},
			"METHOD_NAME":   predictor.TypeMethodName{},
			"FUNCTION_NAME": predictor.FunctionName{},
		}),
	)

	// Пропускаем директивы bash, zsh и т.п.
	args := os.Args
	for i, v := range args {
		if v == appName {
			args = args[i+1:]
			break
		}
	}
	ctx, err := parser.Parse(args)
	parser.FatalIfErrorf(err)

	runArgs := &runContext{
		args:    &cli,
		lookup:  mockLookup,
		logging: logging,
		opts:    genOpts,
	}

	if err := ctx.Run(runArgs); err != nil {
		return err
	}

	return nil
}
