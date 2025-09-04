package client

import (
	"fmt"
	"os"
	"reflect"

	"github.com/pelletier/go-toml/v2"
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
		IdleAfter  string `toml:"idle_after"`
		ViewAfter  string `toml:"view_after"`
		LineOffset string `toml:"line_offset"`
	} `toml:"lsp"`

	LanguageMaps struct {
		URL string `toml:"url"`
	} `toml:"language_maps"`

	Logging struct {
		Level  string `toml:"level"`
		Output string `toml:"output"`
	} `toml:"logging"`
}

func DefaultConfig() *Config {
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

				State:       "{action}",
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
			IdleAfter  string `toml:"idle_after"`
			ViewAfter  string `toml:"view_after"`
			LineOffset string `toml:"line_offset"`
		}{
			IdleAfter:  "5m",
			ViewAfter:  "30s",
			LineOffset: "+1",
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
	config := DefaultConfig()

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		data, err := toml.Marshal(config)
		if err != nil {
			Error("Failed to marshal default config.", map[string]any{
				"error": err,
			})
			return nil, err
		}

		if err := os.WriteFile(configFilePath, data, 0644); err != nil {
			Error("Failed to write default config file.", map[string]any{
				"error": err,
			})
			return nil, err
		}

		Info("Created default config file.", map[string]any{
			"filepath": configFilePath,
		})
	} else {
		data, err := os.ReadFile(configFilePath)
		if err != nil {
			Error("Failed to read config file.", map[string]any{
				"error": err,
			})
			return nil, err
		}

		if err := toml.Unmarshal(data, config); err != nil {
			Error("Failed to unmarshal config file.", map[string]any{
				"error": err,
			})
			return nil, err
		}

		err = fmt.Errorf("config is nil")
		if config == nil {
			Error("Unexpected config file.", map[string]any{
				"error": err,
			})
			return nil, err
		}

		defaultCfg := DefaultConfig()
		loadedCfg := config

		defaultVal := reflect.ValueOf(defaultCfg).Elem()
		loadedVal := reflect.ValueOf(loadedCfg).Elem()

		for i := range defaultVal.NumField() {
			fieldName := defaultVal.Type().Field(i).Name
			defaultField := defaultVal.Field(i)
			loadedField := loadedVal.FieldByName(fieldName)

			if !loadedField.IsValid() {
				Warn("Missing field in config.", map[string]any{
					"fieldName": fieldName,
				})
				return nil, fmt.Errorf("missing field: %s", fieldName)
			}

			if loadedField.Type() != defaultField.Type() {
				Warn("Field has incorrect type.", map[string]any{
					"fieldName":         fieldName,
					"loadedFieldType":   loadedField.Type(),
					"expectedFieldType": defaultField.Type(),
				})
				return nil, fmt.Errorf("field %s has incorrect type: expected %s, got %s", fieldName, defaultField.Type(), loadedField.Type())
			}

			if defaultField.Kind() == reflect.Struct {
				for j := range defaultField.NumField() {
					nestedFieldName := defaultField.Type().Field(j).Name
					nestedDefaultField := defaultField.Field(j)
					nestedLoadedField := loadedField.FieldByName(nestedFieldName)

					if !nestedLoadedField.IsValid() {
						Warn("Missing nested field in config.", map[string]any{
							"parentFieldName":        fieldName,
							"missingNestedFieldName": nestedFieldName,
						})
						return nil, fmt.Errorf("missing nested field: %s.%s", fieldName, nestedFieldName)
					}

					if nestedLoadedField.Type() != nestedDefaultField.Type() {
						Warn("Nested field has incorrect type", map[string]any{
							"parentFieldName":   fieldName,
							"NestedFieldName":   nestedFieldName,
							"fullFieldName":     fmt.Sprintf("%s.%s", fieldName, nestedFieldName),
							"expectedFieldType": nestedDefaultField.Type(),
							"loadedFieldType":   nestedLoadedField.Type(),
						})
						return nil, fmt.Errorf("field %s.%s has incorrect type: expected %s, got %s", fieldName, nestedFieldName, nestedDefaultField.Type(), nestedLoadedField.Type())
					}
				}
			}
		}
	}

	return config, nil
}
