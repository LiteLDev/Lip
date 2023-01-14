// Package localfile deals with ~/.lip directory and its contents
// and ./.lip directory and its contents.
package localfile

import (
	"encoding/base64"
	"errors"
	"os"
)

// Init initializes the ~/.lip and ./.lip directories.
// It should be called before any other functions in this package.
func Init() error {
	// Initialize the ~/.lip directory.
	homeLipDir, err := HomeLipDir()
	if err != nil {
		return err
	}
	cacheDir, err := CacheDir()
	if err != nil {
		return err
	}
	os.MkdirAll(homeLipDir, 0755)
	os.MkdirAll(cacheDir, 0755)

	workspaceLipDir, err := WorkspaceLipDir()
	if err != nil {
		return err
	}
	os.MkdirAll(workspaceLipDir, 0755)

	return nil
}

// CacheDir returns the path to the ~/.lip/cache directory.
func CacheDir() (string, error) {
	homeLipDir, err := HomeLipDir()
	if err != nil {
		return "", err
	}
	cacheDir := homeLipDir + "/cache"
	return cacheDir, nil
}

// GetCachedToothFIlePath returns the path to the cached tooth file.
func GetCachedToothFileName(fullSpecifier string) string {
	// Encode the full specifier with Base64.
	fullSpecifier = base64.StdEncoding.EncodeToString([]byte(fullSpecifier))

	return fullSpecifier + ".tt"
}

// HomeLipDir returns the path to the ~/.lip directory.
func HomeLipDir() (string, error) {
	// Set context.HomeLipDir.
	dirname, err := os.UserHomeDir()
	if err != nil {
		err = errors.New("failed to get user home directory")
		return "", err
	}
	homeLipDir := dirname + "/.lip"
	return homeLipDir, nil
}

func IsCachedToothFileExist(fullSpecifier string) (bool, error) {
	// Get the path to the cached tooth file.
	cachedToothFileName := GetCachedToothFileName(fullSpecifier)

	// Check if the cached tooth file exists.
	cacheDir, err := CacheDir()
	if err != nil {
		return false, err
	}

	cachedToothFilePath := cacheDir + "/" + cachedToothFileName
	if _, err := os.Stat(cachedToothFilePath); err != nil {
		return false, nil
	}

	return true, nil
}

func RecordDir() (string, error) {
	workspaceLipDir, err := WorkspaceLipDir()
	if err != nil {
		return "", err
	}
	recordDir := workspaceLipDir + "/records"
	return recordDir, nil
}

// WorkspaceLipDir returns the path to the ./.lip directory.
func WorkspaceLipDir() (string, error) {
	// Set context.WorkspaceLipDir.
	dirname, err := os.Getwd()
	if err != nil {
		err = errors.New("failed to get current directory")
		return "", err
	}
	workspaceLipDir := dirname + "/.lip"
	return workspaceLipDir, nil
}