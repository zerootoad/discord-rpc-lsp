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

func UpdateDiscordActivity(state, details, currentLang, editor, gitRemoteURL, gitBranchName string, timestamp *time.Time) {
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

	largeImage := "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/" + editor + ".png"
	if editor == "neovim" {
		largeImage = "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/Nvemo.png"
	}
	resp, err := http.Get(largeImage)
	if resp.StatusCode != 200 || err != nil {
		log.Printf("Large image not found, using text icon: %v", err)
		largeImage = "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/text.png"
	} else {
		log.Printf("Large image found: %v", largeImage)
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
		log.Println("Created buttons")
		activity.Buttons = []*client.Button{
			{
				Label: "View Repository",
				Url:   gitRemoteURL,
			},
		}
		activity.Details += " (" + gitBranchName + ")"
	}

	err = client.SetActivity(activity)
	if err != nil {
		log.Printf("Failed to update Rich Presence: %v", err)
	}
}

func ClearDiscordActivity(state, details, editor, gitRemoteURL, gitBranchName string) {
	largeImage := "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/" + editor + ".png"
	if editor == "neovim" {
		largeImage = "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/Nvemo.png"
	}
	resp, err := http.Get(largeImage)
	if resp.StatusCode != 200 || err != nil {
		log.Printf("Large image not found, using text icon: %v", err)
		largeImage = "https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/icons/text.png"
	} else {
		log.Printf("Large image found: %v", largeImage)
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

	err = client.SetActivity(activity)
	if err != nil {
		log.Printf("Failed to update Rich Presence: %v", err)
	}
}
