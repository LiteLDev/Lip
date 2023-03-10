package cmdlipinstall

import (
	"container/list"
	"flag"
	"os"
	"path/filepath"

	cmdlipuninstall "github.com/liteldev/lip/cmd/uninstall"
	"github.com/liteldev/lip/download"
	"github.com/liteldev/lip/localfile"
	"github.com/liteldev/lip/specifiers"
	"github.com/liteldev/lip/tooth/toothfile"
	"github.com/liteldev/lip/tooth/toothrecord"
	"github.com/liteldev/lip/tooth/toothrepo"
	"github.com/liteldev/lip/utils/logger"
	"github.com/liteldev/lip/utils/versions"
)

// FlagDict is a dictionary of flags.
type FlagDict struct {
	helpFlag            bool
	upgradeFlag         bool
	forceReinstallFlag  bool
	yesFlag             bool
	numericProgressFlag bool
	noDependenciesFlag  bool
}

const helpMessage = `
Usage:
  lip install [options] <specifiers>

Description:
  Install tooths from:

  - tooth repositories.
  - local or remote standalone tooth files (with suffix .tth).

Options:
  -h, --help                  Show help.
  --upgrade                   Upgrade the specified tooth to the newest available version.
  --force-reinstall           Reinstall the tooth even if they are already up-to-date.
  -y, --yes                   Assume yes to all prompts and run non-interactively.
  --numeric-progress          Show numeric progress instead of progress bar.
  --no-dependencies            Do not install dependencies.`

// Run is the entry point.
func Run(args []string) {
	var err error

	flagSet := flag.NewFlagSet("install", flag.ExitOnError)

	// Rewrite the default usage message.
	flagSet.Usage = func() {
		logger.Info(helpMessage)
	}

	var flagDict FlagDict
	flagSet.BoolVar(&flagDict.helpFlag, "help", false, "")
	flagSet.BoolVar(&flagDict.helpFlag, "h", false, "")
	flagSet.BoolVar(&flagDict.upgradeFlag, "upgrade", false, "")
	flagSet.BoolVar(&flagDict.forceReinstallFlag, "force-reinstall", false, "")
	flagSet.BoolVar(&flagDict.yesFlag, "yes", false, "")
	flagSet.BoolVar(&flagDict.yesFlag, "y", false, "")
	flagSet.BoolVar(&flagDict.numericProgressFlag, "numeric-progress", false, "")
	flagSet.BoolVar(&flagDict.noDependenciesFlag, "no-dependencies", false, "")
	flagSet.Parse(args)

	// Help flag has the highest priority.
	if flagDict.helpFlag {
		logger.Info(helpMessage)
		return
	}

	// At least one argument is required.
	if flagSet.NArg() == 0 {
		logger.Error("Too few arguments")
		os.Exit(1)
	}

	// 1. Validate the requirement specifier or tooth url/path.
	//    This process will check if the tooth file exists or the tooth url can be accessed
	//    and if the requirement specifier syntax is valid. For requirement specifier, it
	//    will also check if the tooth repository can be accessed via GOPROXY.

	logger.Info("Validating specifiers...")

	// Make requirementSpecifierList.
	var requirementSpecifierList []specifiers.Specifier
	for _, specifierString := range flagSet.Args() {
		logger.Info("  Validating " + specifierString + "...")

		specifier, err := specifiers.New(specifierString)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}

		requirementSpecifierList = append(requirementSpecifierList, specifier)
	}

	// 2. Parse dependency and download tooth files.
	//    This process will maintain an array of tooth files to be downloaded. For each
	//    specifier at the beginning, it will be added to the array. Then, for each
	//    specifier in the array, if it is downloaded, it will be removed from the array.
	//    If it is not downloaded, it will be parsed to get its dependencies and add them
	//    to the array. This process will continue until the array is empty.

	logger.Info("Resolving dependencies and downloading tooths...")

	// An queue of tooth files to be downloaded.
	var specifiersToFetch list.List

	// Add all specifiers to the queue.
	for _, specifier := range requirementSpecifierList {
		specifiersToFetch.PushBack(specifier)
	}

	// An array of downloaded tooth files.
	// Specifier string -> downloaded tooth file path
	downloadedToothFilePathMap := make(map[string]string)

	for specifiersToFetch.Len() > 0 {
		// Get the first specifier to fetch
		specifier := specifiersToFetch.Front().Value.(specifiers.Specifier)
		specifiersToFetch.Remove(specifiersToFetch.Front())

		// If the tooth file of the specifier is already downloaded, skip.
		if _, ok := downloadedToothFilePathMap[specifier.String()]; ok {
			continue
		}

		logger.Info("  Fetching " + specifier.String() + "...")

		var progressBarStyle download.ProgressBarStyleType
		if logger.GetLevel() > logger.InfoLevel {
			progressBarStyle = download.StyleNone
		} else if flagDict.numericProgressFlag {
			progressBarStyle = download.StylePercentageOnly
		} else {
			progressBarStyle = download.StyleDefault
		}

		// Get tooth file
		isCached, downloadedToothFilePath, err := getTooth(specifier, progressBarStyle)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		if isCached {
			logger.Info("    Cached.")
		}

		// Parse the tooth file
		toothFile, err := toothfile.New(downloadedToothFilePath)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}

		// Validate the tooth file.
		toothPath := toothFile.Metadata().ToothPath
		if specifier.Type() == specifiers.RequirementKind &&
			toothPath != specifier.ToothRepo() {
			logger.Error("The tooth path of the downloaded tooth file does not match the requirement specifier")

			// Remove the downloaded tooth file.
			err = os.Remove(downloadedToothFilePath)
			if err != nil {
				logger.Error("Failed to remove the downloaded tooth file: " + err.Error())
			}
			os.Exit(1)
		}

		isToothInstalled, err := toothrecord.IsToothInstalled(toothPath)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		if isToothInstalled {
			if !flagDict.forceReinstallFlag && !flagDict.upgradeFlag {
				logger.Info("    Already installed.")
				continue
			} else if !flagDict.forceReinstallFlag && flagDict.upgradeFlag {
				record, err := toothrecord.Get(toothPath)
				if err != nil {
					logger.Error(err.Error())
					os.Exit(1)
				}
				if record.Version == toothFile.Metadata().Version {
					logger.Info("    Already installed.")
					continue
				} else if versions.GreaterThan(record.Version, toothFile.Metadata().Version) {
					logger.Info("    A newer version already installed.")
					continue
				}
			} else if flagDict.forceReinstallFlag && !flagDict.upgradeFlag {
				logger.Info("    Already installed, reinstalling...")
			} else if flagDict.forceReinstallFlag && flagDict.upgradeFlag {
				logger.Info("    Already installed, reinstalling...")
			}
		}

		// Add the downloaded path to the downloaded tooth files.
		downloadedToothFilePathMap[specifier.String()] = downloadedToothFilePath

		// If the no-dependencies flag is set, skip.
		if !flagDict.noDependenciesFlag {
			logger.Info("    Resolving dependencies...")

			dependencies := toothFile.Metadata().Dependencies

			// Get proper version of each dependency and add them to the queue.
			for toothPath, versionRange := range dependencies {
				logger.Info("      Resolving " + toothPath + "...")

				versionList, err := toothrepo.FetchVersionList(toothPath)
				if err != nil {
					logger.Error(err.Error())
					os.Exit(1)
				}

				// If the dependency is installed, check if the installed version matches the requirement.
				ok, err := toothrecord.IsToothInstalled(toothPath)
				if err != nil {
					logger.Error(err.Error())
					os.Exit(1)
				}
				if ok {
					record, err := toothrecord.New(toothPath)
					if err != nil {
						logger.Error(err.Error())
						os.Exit(1)
					}

					version := record.Version

					var isAnyMatched bool = false
					for _, innerVersionRange := range versionRange {
						var isAllMatched bool = true
						for _, versionMatch := range innerVersionRange {
							if !versionMatch.Match(version) {
								isAllMatched = false
								break
							}
						}

						if isAllMatched {
							isAnyMatched = true
							break
						}
					}

					if isAnyMatched {
						logger.Info("        Installed version " + version.String() + " matches the requirement.")
						continue
					} else {
						versionRangeString := ""
						for i, innerVersionRange := range versionRange {
							if i > 0 {
								versionRangeString += " or "
							}

							versionRangeString += "("

							for j, versionMatch := range innerVersionRange {
								if j > 0 {
									versionRangeString += " and "
								}

								versionRangeString += versionMatch.String()
							}

							versionRangeString += ")"
						}
						logger.Error("The installed version of " + toothPath + "(" + version.String() + ") does not match the requirement of " + specifier.String() + "(" + versionRangeString + ")")
						os.Exit(1)
					}
				}

				isMatched := false
				for _, version := range versionList {
					var isAnyMatched bool = false
					for _, innerVersionRange := range versionRange {
						var isAllMatched bool = true
						for _, versionMatch := range innerVersionRange {
							if !versionMatch.Match(version) {
								isAllMatched = false
								break
							}
						}

						if isAllMatched {
							isAnyMatched = true
							break
						}
					}

					if isAnyMatched {
						// Add the specifier to the queue.
						specifier, err := specifiers.New(toothPath + "@" + version.String())
						if err != nil {
							logger.Error(err.Error())
							os.Exit(1)
						}
						specifiersToFetch.PushBack(specifier)
						isMatched = true
						break
					}
				}

				if !isMatched {
					// If no version is selected, error.
					logger.Error("No version of " + toothPath + " matches the requirement of " + specifier.String())
					os.Exit(1)
				}
			}
		}
	}

	// 3. Deal with force reinstall flag and upgrade flag.
	//    This process will check if the force reinstall flag is set. If it is set, all
	//    installed tooth specified by the specifiers will be reinstalled. If it is not
	//    set, it will check if the upgrade flag is set. If it is set, all installed tooth
	//    specified by the specifiers will be upgraded. If it is not set, all installed
	//    tooth specified by the specifiers will be skipped.

	if flagDict.forceReinstallFlag || flagDict.upgradeFlag {
		if flagDict.forceReinstallFlag && flagDict.upgradeFlag {
			logger.Error("the force-reinstall flag and the upgrade flag cannot be used together")
			os.Exit(1)
		}

		if flagDict.forceReinstallFlag {
			logger.Info("--force-reinstall flag is set, Lip will reinstall all installed tooths specified by the specifiers...")
		} else if flagDict.upgradeFlag {
			logger.Info("--upgrade flag is set, Lip will upgrade all installed tooths specified by the specifiers...")
		}

		logger.Info("Uninstalling tooths to be reinstalled or upgraded...")

		for _, specifier := range requirementSpecifierList {
			// If the specifier is not a requirement specifier, skip.
			if specifier.Type() != specifiers.RequirementKind {
				logger.Error("The specifier " + specifier.String() + " is not a requirement specifier. It cannot be used with the --force-reinstall flag or the --upgrade flag")
				continue
			}

			if _, ok := downloadedToothFilePathMap[specifier.String()]; !ok {
				// Already processed.
				continue
			}

			logger.Info("  Resolving " + specifier.String() + "...")

			toothFile, err := toothfile.New(downloadedToothFilePathMap[specifier.String()])
			if err != nil {
				logger.Error(err.Error())
				os.Exit(1)
			}

			// If the tooth file of the specifier is not installed, skip.
			isInstalled, err := toothrecord.IsToothInstalled(toothFile.Metadata().ToothPath)
			if err != nil {
				logger.Error(err.Error())
				os.Exit(1)
			}
			if !isInstalled {
				continue
			}

			// Get the tooth record.
			recordDir, err := localfile.RecordDir()
			if err != nil {
				logger.Error(err.Error())
				os.Exit(1)
			}

			toothRecordFilePath := filepath.Join(recordDir, localfile.GetRecordFileName(toothFile.Metadata().ToothPath))
			toothRecord, err := toothrecord.NewFromFile(toothRecordFilePath)
			if err != nil {
				logger.Error(err.Error())
				os.Exit(1)
			}

			// Compare the version of the tooth file and the version of the tooth record.
			// If the version of the tooth file is not greater than the version of the tooth
			// record, skip.
			if !flagDict.forceReinstallFlag && !versions.GreaterThan(toothFile.Metadata().Version, toothRecord.Version) {
				continue
			}

			// If the tooth file of the specifier is installed, uninstall it.
			logger.Info("    Uninstalling " + toothFile.Metadata().ToothPath + "...")

			possessionList := toothFile.Metadata().Possession
			recordFileName := localfile.GetRecordFileName(toothFile.Metadata().ToothPath)

			err = cmdlipuninstall.Uninstall(recordFileName, possessionList, flagDict.yesFlag)
			if err != nil {
				logger.Error(err.Error())
				os.Exit(1)
			}
		}

	}

	// 4. Install tooth files.
	//    This process will install all downloaded tooth files. If the tooth file is
	//    already installed, it will be skipped. If the tooth file is not installed, it
	//    will be installed. If the tooth file is installed but the version is different,
	//    it will be upgraded or reinstalled according to the flags.

	logger.Info("Installing tooths...")

	// Store downloaded tooth files in an array.
	downloadedToothFileList := make([]toothfile.ToothFile, 0)
	manuallyInstalledToothFilePathList := make([]string, 0)
	for specifierString, downloadedToothFilePath := range downloadedToothFilePathMap {
		toothFile, err := toothfile.New(downloadedToothFilePath)
		if err != nil {
			logger.Error("Failed to open the downloaded tooth file " + downloadedToothFilePath + ": " + err.Error())
			os.Exit(1)
		}

		// If specifierString is in the requirementSpecifierList, it is a manually
		// installed tooth file. Otherwise, it is a dependency tooth file.
		isManuallyInstalled := false
		for _, requirementSpecifier := range requirementSpecifierList {
			if requirementSpecifier.String() == specifierString {
				isManuallyInstalled = true
				break
			}
		}

		downloadedToothFileList = append(downloadedToothFileList, toothFile)

		if isManuallyInstalled {
			manuallyInstalledToothFilePathList = append(manuallyInstalledToothFilePathList, toothFile.FilePath())
		}
	}

	// Topological sort the array in descending order.
	downloadedToothFileList, err = sortToothFiles(downloadedToothFileList)
	if err != nil {
		logger.Error("Failed to sort the downloaded tooth files: " + err.Error())
		os.Exit(1)
	}

	for _, toothFile := range downloadedToothFileList {
		logger.Info("  Resolving " + toothFile.Metadata().ToothPath + "@" + toothFile.Metadata().Version.String() + "...")

		// If the tooth file is already installed, skip.
		isInstalled, err := toothrecord.IsToothInstalled(toothFile.Metadata().ToothPath)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		if isInstalled {
			logger.Info("    " + toothFile.Metadata().ToothPath + " is already installed.")
			continue
		}

		// Install the tooth file.
		logger.Info("    Installing " + toothFile.Metadata().ToothPath + "@" +
			toothFile.Metadata().Version.String() + "...")

		isManuallyInstalled := false
		for _, manuallyInstalledToothFilePath := range manuallyInstalledToothFilePathList {
			if toothFile.FilePath() == manuallyInstalledToothFilePath {
				isManuallyInstalled = true
				break
			}
		}

		err = install(toothFile, isManuallyInstalled, flagDict.yesFlag)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
	}

	logger.Info("Successfully installed all tooth files.")
}
