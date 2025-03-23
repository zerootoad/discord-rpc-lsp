package client

import (
	"net/http"
	"time"

	"github.com/hugolgst/rich-go/client"
)

func Login(applicationID string) error {
	return client.Login(applicationID)
}

func Logout() {
	client.Logout()
}

func UpdateDiscordActivity(state, details, currentLang, editor, gitRemoteURL, gitBranchName string, timestamp *time.Time) error {
	smallImage := ""
	smallText := ""
	if state != "Idling" {
		smallImage = "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/" + currentLang + ".png"
		resp, err := http.Get(smallImage)
		if resp.StatusCode != 200 || err != nil {
			smallImage = "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/text.png"
		}

		defer resp.Body.Close()

		smallText = "Coding in " + currentLang
	}

	largeImage := "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/" + editor + ".png"
	if editor == "neovim" {
		largeImage = "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/Nvemo.png"
	}
	resp, err := http.Get(largeImage)
	if resp.StatusCode != 200 || err != nil {
		largeImage = "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/text.png"
	}
	defer resp.Body.Close()

	activity := client.Activity{}
	if currentLang == "" {
		activity = client.Activity{
			State:      state,
			Details:    details,
			LargeImage: largeImage,
			LargeText:  editor,
			Timestamps: &client.Timestamps{
				Start: timestamp,
			},
		}
	} else {
		activity = client.Activity{
			State:      state,
			Details:    details,
			LargeImage: largeImage,
			LargeText:  editor,
			SmallImage: smallImage,
			SmallText:  smallText,
			Timestamps: &client.Timestamps{
				Start: timestamp,
			},
		}
	}

	if gitRemoteURL != "" {
		activity.Buttons = []*client.Button{
			{
				Label: "View Repository",
				Url:   gitRemoteURL,
			},
		}
		activity.Details += " (" + gitBranchName + ")"
	}

	return client.SetActivity(activity)
}

func ClearDiscordActivity(state, details, editor, gitRemoteURL, gitBranchName string) error {
	largeImage := "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/" + editor + ".png"
	if editor == "neovim" {
		largeImage = "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/Nvemo.png"
	}
	resp, err := http.Get(largeImage)
	if resp.StatusCode != 200 || err != nil {
		largeImage = "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/text.png"
	}
	defer resp.Body.Close()

	now := time.Now()
	activity := client.Activity{
		State:      state,
		Details:    details,
		LargeImage: largeImage,
		LargeText:  editor,
		Timestamps: &client.Timestamps{
			Start: &now,
		},
	}

	if gitRemoteURL != "" {
		activity.Buttons = []*client.Button{
			{
				Label: "View Repository",
				Url:   gitRemoteURL,
			},
		}
		activity.Details += " (" + gitBranchName + ")"
	}

	return client.SetActivity(activity)
}
