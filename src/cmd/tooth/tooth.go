package cmdliptooth

import (
	"flag"

	cmdliptoothinit "github.com/liteldev/lip/cmd/tooth/init"
	"github.com/liteldev/lip/utils/logger"
)

// FlagDict is a dictionary of flags.
type FlagDict struct {
	helpFlag bool
}

const helpMessage = `
Usage:
  lip tooth [options]
  lip tooth <command> [subcommand options] ...

Commands:
  init                        Initialize and writes a new tooth.json file in the current directory.

Options:
  -h, --help                  Show help.`

// Run is the entry point.
func Run(args []string) {
	// If there is a subcommand, run it and exit.
	if len(args) >= 1 {
		switch args[0] {
		case "init":
			cmdliptoothinit.Run(args[1:])
			return
		}
	}

	flagSet := flag.NewFlagSet("tooth", flag.ExitOnError)

	// Rewrite the default usage message.
	flagSet.Usage = func() {
		logger.Info(helpMessage)
	}

	var flagDict FlagDict

	flagSet.BoolVar(&flagDict.helpFlag, "help", false, "")
	flagSet.BoolVar(&flagDict.helpFlag, "h", false, "")

	flagSet.Parse(args)

	// Help flag has the highest priority.
	if flagDict.helpFlag {
		logger.Info(helpMessage)
		return
	}

	// If there is no subcommand, print help message and exit.
	logger.Info(helpMessage)
}
