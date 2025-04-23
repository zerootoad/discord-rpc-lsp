package client

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hugolgst/rich-go/client"
	log "github.com/sirupsen/logrus"
	"github.com/zerootoad/discord-rpc-lsp/utils"
)

var (
	debouncer = utils.NewDebouncer(5 * time.Second)
)

func Login(applicationID string) error {
	return client.Login(applicationID)
}

func Logout() {
	client.Logout()
}

func replacePlaceholders(s string, placeholders map[string]string) string {
	for placeholder, value := range placeholders {
		s = strings.Replace(s, placeholder, value, -1)
	}
	return s
}

func updateActivityConfig(config *Config, placeholders map[string]string) ActivityConfig {
	newActivity := ActivityConfig{
		State:      replacePlaceholders(config.Discord.Activity.State, placeholders),
		Details:    replacePlaceholders(config.Discord.Activity.Details, placeholders),
		LargeImage: replacePlaceholders(config.Discord.Activity.LargeImage, placeholders),
		LargeText:  replacePlaceholders(config.Discord.Activity.LargeText, placeholders),
		SmallImage: replacePlaceholders(config.Discord.Activity.SmallImage, placeholders),
		SmallText:  replacePlaceholders(config.Discord.Activity.SmallText, placeholders),
	}

	return newActivity
}

func getImageURL(url string, defaultURL string) string {
	if url == "" {
		resp, err := http.Get(defaultURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			return "https://raw.githubusercontent.com/zerootoad/discord-rpc-lsp/refs/heads/main/assets/icons/text.png"
		}
		defer resp.Body.Close()
		return defaultURL
	}

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return defaultURL
	}
	defer resp.Body.Close()
	return url
}

func UpdateDiscordActivity(config *Config, tempaction, filename, workspace, currentLang, editor, gitRemoteURL, gitBranchName string, timestamp *time.Time) error {
	if strings.Contains(workspace, os.TempDir()) {
		workspace = editor
	}

	placeholders := map[string]string{
		"{filename}":  filename,
		"{workspace}": workspace,
		"{editor}":    editor,
		"{language}":  currentLang,
	}

	action := replacePlaceholders(tempaction, placeholders)
	placeholders["{action}"] = action

	tempActivity := updateActivityConfig(config, placeholders)

	smallImage := getImageURL(tempActivity.SmallImage, replacePlaceholders("https://raw.githubusercontent.com/zerootoad/discord-rpc-lsp/refs/heads/main/assets/icons/{language}.png", placeholders))
	largeImage := getImageURL(tempActivity.LargeImage, replacePlaceholders("https://raw.githubusercontent.com/zerootoad/discord-rpc-lsp/refs/heads/main/assets/icons/{editor}.png", placeholders))
	if editor == "neovim" && strings.Contains(largeImage, "zerootoad") {
		largeImage = "https://raw.githubusercontent.com/zerootoad/discord-rpc-lsp/refs/heads/main/assets/icons/Nvemo.png"
	}

	if currentLang == "" {
		smallImage = ""
		tempActivity.SmallText = ""
	}

	activity := client.Activity{
		State:      tempActivity.State,
		Details:    tempActivity.Details,
		LargeImage: largeImage,
		LargeText:  tempActivity.LargeText,
		SmallImage: smallImage,
		SmallText:  tempActivity.SmallText,
	}

	switch config.Discord.LargeUse {
	case "language":
		activity.LargeImage = smallImage
		activity.LargeText = tempActivity.SmallText
	case "editor":
		activity.LargeImage = largeImage
		activity.LargeText = tempActivity.LargeText
	}

	switch config.Discord.SmallUse {
	case "language":
		activity.SmallImage = smallImage
		activity.SmallText = tempActivity.SmallText
	case "editor":
		activity.SmallImage = largeImage
		activity.SmallText = tempActivity.LargeText
	}

	if config.Discord.Activity.Timestamp {
		activity.Timestamps = &client.Timestamps{
			Start: timestamp,
		}
	}

	if gitRemoteURL != "" && config.Git.GitInfo {
		activity.Buttons = []*client.Button{
			{
				Label: "View Repository",
				Url:   gitRemoteURL,
			},
		}
		activity.Details += " (" + gitBranchName + ")"
	}

	var err error
	debouncer.Debounce(func() {
		log.Info("Updating discord activity")
		err = client.SetActivity(activity)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to update Discord activity")
		}
	})
	return err
}

func ClearDiscordActivity(config *Config, action, filename, workspace, editor, gitRemoteURL, gitBranchName string) error {
	placeholders := map[string]string{
		"{action}":    action,
		"{filename}":  filename,
		"{workspace}": workspace,
		"{editor}":    editor,
	}

	tempActivity := updateActivityConfig(config, placeholders)

	largeImage := getImageURL(tempActivity.LargeImage, replacePlaceholders("https://raw.githubusercontent.com/zerootoad/discord-rpc-lsp/refs/heads/main/assets/icons/{editor}.png", placeholders))
	if editor == "neovim" && strings.Contains(largeImage, "zerootoad") {
		largeImage = "https://raw.githubusercontent.com/zerootoad/discord-rpc-lsp/refs/heads/main/assets/icons/Nvemo.png"
	}

	now := time.Now()
	activity := client.Activity{
		State:      tempActivity.State,
		Details:    tempActivity.Details,
		LargeImage: largeImage,
		LargeText:  tempActivity.LargeText,
		Timestamps: &client.Timestamps{
			Start: &now,
		},
	}

	if gitRemoteURL != "" && config.Git.GitInfo {
		activity.Buttons = []*client.Button{
			{
				Label: "View Repository",
				Url:   gitRemoteURL,
			},
		}
		activity.Details += " (" + gitBranchName + ")"
	}

	var err error
	debouncer.Debounce(func() {
		log.Info("Clear discord activity")
		err = client.SetActivity(activity)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to clear Discord activity")
		}
	})
	return err
}
