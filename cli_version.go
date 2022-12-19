package ttgenlib

import (
	"runtime/debug"

	"github.com/sirkon/message"
)

// versionCommand show version command.
type versionCommand struct{}

// Run запуск вывода версии
func (versionCommand) Run(ctx *runContext) error {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		message.Warning(
			"WARNING: you are using a version compiled with modules disabled, this is not the way it supposed to be",
		)
	} else {
		message.Info(globalVarAppName, "version", info.Main.Version)
	}

	return nil
}
