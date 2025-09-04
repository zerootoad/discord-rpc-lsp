package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Throttler struct {
	interval time.Duration
	lastRun  time.Time
	mu       sync.Mutex
}

func NewThrottler(interval time.Duration) *Throttler {
	return &Throttler{
		interval: interval,
	}
}

func (t *Throttler) Run(f func()) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	if now.Sub(t.lastRun) < t.interval {
		return
	}

	t.lastRun = now
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

func EvalOffset(expr string) int {
	expr = strings.TrimSpace(expr)
	expr = strings.ReplaceAll(expr, "+", " + ")
	expr = strings.ReplaceAll(expr, "-", " - ")
	expr = strings.Join(strings.Fields(expr), " ")

	parts := strings.Split(expr, " ")
	if len(parts) == 1 {
		val, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0
		}
		return val
	}

	if len(parts) == 3 {
		left, err1 := strconv.Atoi(parts[0])
		op := parts[1]
		right, err2 := strconv.Atoi(parts[2])
		if err1 != nil || err2 != nil {
			return 0
		}

		switch op {
		case "+":
			return left + right
		case "-":
			return left - right
		}
	}

	return 0
}
