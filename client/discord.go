package client

import (
	"log"
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

func UpdateDiscordActivity(state, details, currentLang, editor, gitRemoteURL, gitBranchName string) {
	smallImage := ""
	smallText := ""
	if state != "Idling" {
		smallImage = "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/" + currentLang + ".png"
		resp, err := http.Get(smallImage)
		if resp.StatusCode != 200 || err != nil {
			log.Printf("Small image not found, using text icon: %v", err)
			smallImage = "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/text.png"
		}

		defer resp.Body.Close()

		smallText = "Coding in " + currentLang
	}

	now := time.Now()
	activity := client.Activity{
		State:      state,
		Details:    details,
		LargeImage: "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/" + editor + ".png",
		LargeText:  editor,
		SmallImage: smallImage,
		SmallText:  smallText,
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

	err := client.SetActivity(activity)
	if err != nil {
		log.Printf("Failed to update Rich Presence: %v", err)
	}
}
