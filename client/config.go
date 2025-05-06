package client

import (
	"fmt"
	"os"
	"reflect"

	"github.com/pelletier/go-toml/v2"
	log "github.com/sirupsen/logrus"
)

type ActivityConfig struct {
	IdleAction  string `toml:"idle_action"`
	ViewAction  string `toml:"view_action"`
	EditAction  string `toml:"edit_action"`
	State       string `toml:"state"`
	Details     string `toml:"details"`
	LargeImage  string `toml:"large_image"`
	LargeText   string `toml:"large_text"`
	SmallImage  string `toml:"small_image"`
	SmallText   string `toml:"small_text"`
	Timestamp   bool   `toml:"timestamp"`
	EditingInfo bool   `toml:"editing_info"`
}

type Config struct {
	Discord struct {
		ApplicationID string         `toml:"application_id"`
		SmallUse      string         `toml:"small_usage"`
		LargeUse      string         `toml:"large_usage"`
		RetryAfter    string         `toml:"retry_after"`
		Activity      ActivityConfig `toml:"activity"`
	} `toml:"discord"`

	Git struct {
		GitInfo bool `toml:"git_info"`
	} `toml:"git"`

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
				IdleAction: "Idle in {editor}",
				ViewAction: "Viewing {filename}",
				EditAction: "Editing {filename}",

				State:       "{action} {filename}",
				Details:     "In {workspace}",
				LargeImage:  "",
				LargeText:   "{editor}",
				SmallImage:  "",
				SmallText:   "Coding in {language}",
				Timestamp:   true,
				EditingInfo: true,
			},
		},
		Git: struct {
			GitInfo bool `toml:"git_info"`
		}{GitInfo: true},
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
	}

	return config, nil
}
