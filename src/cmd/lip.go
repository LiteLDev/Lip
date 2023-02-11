// Package cmdlip is the entry point of the lip command.
package cmdlip

import (
	"flag"
	"os"
	"path/filepath"

	cmdlipcache "github.com/liteldev/lip/cmd/cache"
	cmdlipexec "github.com/liteldev/lip/cmd/exec"
	cmdlipinstall "github.com/liteldev/lip/cmd/install"
	cmdliplist "github.com/liteldev/lip/cmd/list"
	cmdlipshow "github.com/liteldev/lip/cmd/show"
	cmdliptooth "github.com/liteldev/lip/cmd/tooth"
	cmdlipuninstall "github.com/liteldev/lip/cmd/uninstall"
	"github.com/liteldev/lip/context"
	"github.com/liteldev/lip/utils/logger"
)

// FlagDict is a dictionary of flags.
type FlagDict struct {
	helpFlag    bool
	versionFlag bool
}

const helpMessage = `
Usage:
  lip [options]
  lip <command> [subcommand options] ...

Commands:
  cache                       Inspect and manage Lip's cache.
  exec                        Execute a Lip tool.
  install                     Install a tooth.
  list                        List installed tooths.
  show                        Show information about installed tooths.
  tooth                       Maintain a tooth.
  uninstall                   Uninstall a tooth.

Options:
  -h, --help                  Show help.
  -V, --version               Show version and exit.`

const versionMessage = "Lip %s from %s"

// Run is the entry point of the lip command.
func Run(args []string) {
	// Initialize context
	context.Init()

	// If there is a subcommand, run it and exit.
	if len(args) >= 1 {
		switch args[0] {
		case "cache":
			cmdlipcache.Run(args[1:])
			return
		case "exec":
			cmdlipexec.Run(args[1:])
			return
		case "install":
			cmdlipinstall.Run(args[1:])
			return
		case "list":
			cmdliplist.Run(args[1:])
			return
		case "show":
			cmdlipshow.Run(args[1:])
			return
		case "tooth":
			cmdliptooth.Run(args[1:])
			return
		case "uninstall":
			cmdlipuninstall.Run(args[1:])
			return
		}
	}

	flagSet := flag.NewFlagSet("lip", flag.ExitOnError)

	// Rewrite the default usage message.
	flagSet.Usage = func() {
		logger.Info(helpMessage)
	}

	// Parse flags.

	var flagDict FlagDict

	flagSet.BoolVar(&flagDict.helpFlag, "help", false, "")
	flagSet.BoolVar(&flagDict.helpFlag, "h", false, "")

	flagSet.BoolVar(&flagDict.versionFlag, "version", false, "")
	flagSet.BoolVar(&flagDict.versionFlag, "V", false, "")

	flagSet.Parse(args)

	// Help flag has the highest priority.
	if flagDict.helpFlag {
		logger.Info(helpMessage)
		return
	}

	if flagDict.versionFlag {
		exPath, _ := filepath.Abs(os.Args[0])
		logger.Info(versionMessage, context.Version.String(), exPath)
		return
	}

	// If there is no flag, print help message and exit.
	logger.Info(helpMessage)
}
