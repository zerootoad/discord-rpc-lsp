package utils

import (
	"os"
	"path/filepath"
	"runtime"
)

func GetUserHomeDir() string {
	if runtime.GOOS == "windows" {
		roamingPath := os.Getenv("APPDATA")
		if roamingPath != "" {
			return roamingPath
		}

		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func GetFileName(uri string) string {
	return filepath.Base(uri)
}

func GetFileExtension(uri string) string {
	return filepath.Ext(uri)
}
