package ttgenlib

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/sirkon/errors"
	"github.com/sirkon/message"
	"github.com/willabides/kongplete"
)

// Run table tests generation with a given names and handlers.
func Run(appName string) error {
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

	if cli.Completions {
		var c kongplete.InstallCompletions
		if err := c.BeforeApply(ctx); err != nil {
			return errors.Wrap(err, "install completions")
		}

		message.Info("completions installed")
		return nil
	}

	if err := validateSubject(cli); err != nil {
		parser.FatalIfErrorf(err)
	}

	return nil
}

func validateSubject(cli cliArgs) error {
	switch {
	case cli.Type != "":
		switch {
		case cli.Method == "":
			return errors.New("missing or empty method name")
		case cli.Function != "":
			return errors.New("type.method and function parameters cannot be used at once")
		}
	case cli.Method != "":
		// Получается что название метода дано, а имя типа – пустое. Это недопустимая ситуация.
		return errors.New("type name must not be empty")
	case cli.Function == "":
		return errors.New("missing type and method or function name")
	}

	return nil
}
