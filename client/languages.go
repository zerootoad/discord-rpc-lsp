package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/zerootoad/discord-rpc-lsp/utils"
)

type LangMaps struct {
	RegexMap map[string]string `json:"RegexMap"`
	ExtMap   map[string]string `json:"ExtMap"`
}

func LoadLangMaps(url string) (LangMaps, error) {
	resp, err := http.Get(url)
	if err != nil {
		return LangMaps{}, fmt.Errorf("error fetching JSON from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return LangMaps{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var langMaps LangMaps
	err = json.NewDecoder(resp.Body).Decode(&langMaps)
	if err != nil {
		return LangMaps{}, fmt.Errorf("error decoding JSON: %w", err)
	}

	return langMaps, nil
}

func (l *LangMaps) GetLanguage(fileName string) string {
	ext := utils.GetFileExtension(fileName)

	if lang, ok := l.ExtMap[ext]; ok {
		return lang
	}

	for pattern, lang := range l.RegexMap {
		re, err := regexp.Compile(pattern)
		if err != nil {
			fmt.Printf("Error compiling regex pattern '%s': %v\n", pattern, err)
			continue
		}

		if re.MatchString(fileName) {
			return lang
		}
	}

	return ""
}
