package utils

import (
	"path/filepath"
)

func GetFileName(uri string) string {
	return filepath.Base(uri)
}

func GetFileExtension(uri string) string {
	return filepath.Ext(uri)
}
