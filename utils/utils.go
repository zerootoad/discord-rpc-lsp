package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type Debouncer struct {
	lastUpdate time.Time
	delay      time.Duration
	mu         sync.Mutex
}

func NewDebouncer(delay time.Duration) *Debouncer {
	return &Debouncer{
		delay: delay,
	}
}

func (d *Debouncer) Debounce(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	if now.Sub(d.lastUpdate) < d.delay {
		return
	}

	d.lastUpdate = now
	f()
}

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
