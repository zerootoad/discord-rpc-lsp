package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pelletier/go-toml/v2"
	log "github.com/sirupsen/logrus"
)

type ActivityConfig struct {
	State      string `toml:"state"`
	Details    string `toml:"details"`
	LargeImage string `toml:"large_image"`
	LargeText  string `toml:"large_text"`
	SmallImage string `toml:"small_image"`
	SmallText  string `toml:"small_text"`
	Timestamp  bool   `toml:"timestamp"`
}

type Config struct {
	Discord struct {
		ApplicationID string         `toml:"application_id"`
		SmallUse      string         `toml:"small_usage"`
		LargeUse      string         `toml:"large_usage"`
		RetryAfter    string         `toml:"retry_after"`
		Activity      ActivityConfig `toml:"activity"`
	} `toml:"discord"`

	Lsp struct {
		Timeout string `toml:"timeout"`
	} `toml:"lsp"`

	LanguageMaps struct {
		URL string `toml:"url"`
	} `toml:"language_maps"`

	Logging struct {
		Level  string `toml:"level"`
		Output string `toml:"output"`
	} `toml:"logging"`
}

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

func defaultConfig() *Config {
	return &Config{
		Discord: struct {
			ApplicationID string         `toml:"application_id"`
			SmallUse      string         `toml:"small_usage"`
			LargeUse      string         `toml:"large_usage"`
			RetryAfter    string         `toml:"retry_after"`
			Activity      ActivityConfig `toml:"activity"`
		}{
			ApplicationID: "",
			SmallUse:      "language",
			LargeUse:      "editor",
			RetryAfter:    "1m",
			Activity: ActivityConfig{
				State:      "{action} {filename}",
				Details:    "in {workspace}",
				LargeImage: "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/main/assets/icons/{editor}.png",
				LargeText:  "{editor}",
				SmallImage: "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/main/assets/icons/{language}.png",
				SmallText:  "Coding in {language}",
				Timestamp:  true,
			},
		},
		Lsp: struct {
			Timeout string `toml:"timeout"`
		}{
			Timeout: "5m",
		},
		LanguageMaps: struct {
			URL string `toml:"url"`
		}{
			URL: "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/main/assets/languages.json",
		},
		Logging: struct {
			Level  string `toml:"level"`
			Output string `toml:"output"`
		}{
			Level:  "info",
			Output: "file",
		},
	}
}

func LoadConfig(configFilePath string) (*Config, error) {
	config := defaultConfig()

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		data, err := toml.Marshal(config)
		if err != nil {
			log.Errorf("Failed to marshal default config: %v", err)
			return nil, err
		}

		if err := os.WriteFile(configFilePath, data, 0644); err != nil {
			log.Errorf("Failed to write default config file: %v", err)
			return nil, err
		}

		log.Infof("Created default config file at: %s", configFilePath)
	} else {
		data, err := os.ReadFile(configFilePath)
		if err != nil {
			log.Errorf("Failed to read config file: %v", err)
			return nil, err
		}

		if err := toml.Unmarshal(data, config); err != nil {
			log.Errorf("Failed to unmarshal config file: %v", err)
			return nil, err
		}

		defaultConfig := defaultConfig()
		defaultVal := reflect.ValueOf(defaultConfig).Elem()
		loadedVal := reflect.ValueOf(config).Elem()

		for i := range defaultVal.NumField() {
			fieldName := defaultVal.Type().Field(i).Name
			defaultField := defaultVal.Field(i)
			loadedField := loadedVal.FieldByName(fieldName)

			if !loadedField.IsValid() {
				log.Errorf("Missing field in config: %s", fieldName)
				return nil, fmt.Errorf("missing field: %s", fieldName)
			}

			if loadedField.Type() != defaultField.Type() {
				log.Errorf("Field %s has incorrect type: expected %s, got %s", fieldName, defaultField.Type(), loadedField.Type())
				return nil, fmt.Errorf("field %s has incorrect type: expected %s, got %s", fieldName, defaultField.Type(), loadedField.Type())
			}

			if defaultField.Kind() == reflect.Struct {
				for j := range defaultField.NumField() {
					nestedFieldName := defaultField.Type().Field(j).Name
					nestedDefaultField := defaultField.Field(j)
					nestedLoadedField := loadedField.FieldByName(nestedFieldName)

					if !nestedLoadedField.IsValid() {
						log.Errorf("Missing nested field in config: %s.%s", fieldName, nestedFieldName)
						return nil, fmt.Errorf("missing nested field: %s.%s", fieldName, nestedFieldName)
					}

					if nestedLoadedField.Type() != nestedDefaultField.Type() {
						log.Errorf("Field %s.%s has incorrect type: expected %s, got %s", fieldName, nestedFieldName, nestedDefaultField.Type(), nestedLoadedField.Type())
						return nil, fmt.Errorf("field %s.%s has incorrect type: expected %s, got %s", fieldName, nestedFieldName, nestedDefaultField.Type(), nestedLoadedField.Type())
					}
				}
			}
		}

		retryAfter, err := ParseDuration(config.Discord.RetryAfter)
		if err != nil {
			log.Errorf("Failed to parse retry_after duration: %v", err)
			return nil, err
		}
		config.Discord.RetryAfter = retryAfter.String()

		timeout, err := ParseDuration(config.Lsp.Timeout)
		if err != nil {
			log.Errorf("Failed to parse timeout duration: %v", err)
			return nil, err
		}
		config.Lsp.Timeout = timeout.String()
	}

	return config, nil
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

func ParseDuration(s string) (time.Duration, error) {
	if len(s) == 0 {
		return 0, fmt.Errorf("empty duration string")
	}

	numStr := ""
	i := 0
	for ; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			numStr += string(s[i])
		} else {
			break
		}
	}

	if numStr == "" {
		return 0, fmt.Errorf("no numeric part in duration string")
	}

	unit := s[i:]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, err
	}

	switch strings.ToLower(unit) {
	case "ns":
		return time.Duration(num) * time.Nanosecond, nil
	case "us", "µs":
		return time.Duration(num) * time.Microsecond, nil
	case "ms":
		return time.Duration(num) * time.Millisecond, nil
	case "s":
		return time.Duration(num) * time.Second, nil
	case "m":
		return time.Duration(num) * time.Minute, nil
	case "h":
		return time.Duration(num) * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown duration unit: %s", unit)
	}
}
