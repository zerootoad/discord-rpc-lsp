# Discord Rich Presence LSP

A Language Server Protocol (LSP) to share what you're coding on Discord. This LSP integrates with your editor to display your current coding activity (file, language, and more) as a rich presence on Discord. 

![image](https://github.com/user-attachments/assets/3089b1ab-0f04-46d3-ae59-ed6207a853f4)
![image](https://github.com/user-attachments/assets/ddbd5f14-65f2-4ab0-a0c5-0b5eb3b36795)

---

## Features

- Displays the file you're currently editing.
- Shows the programming language you're using.
- Includes Git repository information (branch and remote URL).
- Customizable rich presence with editor-specific icons.
- Supports multiple editors via LSP.

---

## TODO

- [x] Fix idle state not resetting timer and showing past edited file for some reason. (possibly fixed, please report this issue if u encounter it)

![image](https://github.com/user-attachments/assets/dbf7da91-8063-4d6c-9e74-31f9d50ed082)
- [ ] Push `go.mod` file and [create tagged releases in github](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository) eg. v1.0.0
- [ ] Improve the project code overall. (really horrid atm)
- [ ] Improve customization. (wip, added ability to change the action placeholder and better image handling)
- [ ] Add diagnostics to the discord activity, best guess (zk way): [refreshDiagnosticsOfDocument](https://github.com/zk-org/zk/blob/68e6b70eaefdf8344065fcec39d5419dc80d6a02/internal/adapter/lsp/server.go#L556)

---
## Installation

### Arch Linux (AUR)

Install the [`discord-rpc-lsp-git`](https://aur.archlinux.org/packages/discord-rpc-lsp-git) from the [AUR](https://aur.archlinux.org/).

```
yay -S discord-rpc-lsp-git
```


## Build Steps

### Prerequisites

- Go (version 1.21 or higher)
- Git
- Discord (with Rich Presence enabled)

### Steps

1. **Clone the Repository**

   ```bash
   git clone https://github.com/zerootoad/discord-rpc-lsp.git
   cd discord-rpc-lsp
   ```
2. **Initialize Go module**

   ```bash
   go mod init github.com/zerootoad/discord-rpc-lsp
   ```
   
2. **Download Dependencies**

   ```bash
   go mod tidy
   ```

3. **Build the Project**

   ```bash
   go build
   ```

4. **Run the LSP Server**

   ```bash
   ./discord-rpc-lsp
   ```

---

## Adding to Editors

### Supported Editors

This LSP works with any editor that supports the Language Server Protocol (LSP). Below are instructions for a few popular editors.

---

### **Visual Studio Code (VS Code) WIP**

---

### **Neovim**

1. Install a plugin like [nvim-lspconfig](https://github.com/neovim/nvim-lspconfig) if you don't already have it.
2. Add the following configuration to your `init.lua`:

   ```lua
   local lspconfig = require('lspconfig')
   local configs = require('lspconfig.configs')


    configs.discord_rpc = {
        default_config = {
            cmd = { "path/to/discord-rpc-lsp" },
            filetypes = {"*"}, -- Add relevant filetypes if needed
            root_dir = function(fname)
                return lspconfig.util.root_pattern('.git')(fname) or vim.fn.getcwd()
            end,
            settings = {},
        },
    }
   ```

3. Replace `path/to/discord-rpc-lsp` with the actual path to the built binary.

---

### **Helix Editor**

1. Open your `languages.toml` file (usually located at `~/.config/helix/languages.toml` for linux).
2. Add the following configuration:

   ```toml
   [language-server.discord-rpc]
   command = "path/to/discord-rpc-lsp"
   ```

3. Replace `path/to/discord-rpc-lsp` with the actual path to the built binary.
4. Add `"discord-rpc"` for the choosen languages:
   ```toml
   [[language]]
   name = "go" # or any language of choice
   language-servers = [ "discord-rpc" ]
   ```


---

### **Other Editors**

For other editors, refer to their documentation on how to configure LSP servers. The process is similar: specify the path to the `discord-rpc-lsp` binary in your editor's LSP configuration.

---

## Configuration

Configuration is done by editing the `config.toml` file located in the configuration directory. The configuration directory is automatically created in the following locations based on your operating system:
- **Unix-based systems (Linux, macOS):** `~/.discord-rpc-lsp/`
- **Windows:** `%APPDATA%\Roaming\.discord-rpc-lsp\`

*Make sure to make reference to the discord rich presence documentation for the fields.* [discord rpc docs](https://discord.com/developers/docs/rich-presence/using-with-the-game-sdk#understanding-rich-presence-data)

By default, if the config.toml file does not exist, it will be created with the following default values:
```toml
[discord]
# Custom Discord Application ID for the Rich Presence.
# This is optional, as the lsp handles it based on the editor being used.
application_id = ''

# Determines what is displayed in the small icon.
# Valid values: "language" or "editor".
small_usage = 'language'

# Determines what is displayed in the large icon.
# Valid values: "language" or "editor".
large_usage = 'editor'

# retry_after is the duration to wait before retrying in case it fails to create the discord rpc client.
# Must be a valid duration string (e.g., "1m", "30s").
retry_after = '1m'

[discord.activity]
# The discord activity is customizable via placeholders.
# 
# List of avaible placeholder:
# {action} : holds the action being executed, can be customized below.
# {filename} : holds the name of current file.
# {workspace} : holds the workspace name.
# {editor} : holds the editor name (e.g., "helix", "neovim")
# {language} : holds the language name of the current file.

# These 3 fields define the {action} placeholder based on the current action.
idle_action = 'Idle in {workspace}'
view_action = 'Viewing {filename}'
edit_action = 'Editing {filename}'

# state is the first line of the activity status.
state = '{action}'

# Details hold the current workspace.
details = 'In {workspace}'

# OPTIONAL: field only fill it if u would like to overwrite the default picked one. (MUST BE A URL TAKING TO THE IMAGE)
large_image = ''

# Large icon text for when u hover over it.
large_text = '{editor}'

# OPTIONAL: field only fill it if u would like to overwrite the default picked one. (MUST BE A URL TAKING TO THE IMAGE)
small_image = ''

# Small icon text for when u hover over it.
small_text = 'Coding in {language}'

# If true, the time since the activity started will be shown.
timestamp = true

# If true, additional information on the file being edited will be shown
editing_info = true

[git]
# If true, will show the repository and branch information
git_info = true

[lsp]
# The duration after which the LSP will enable idling if no activity is detected.
# Must be a valid duration string (e.g., "5m", "30s").
timeout = '5m'

[language_maps]
# The URL to a JSON file containing mappings of file extensions to programming languages.
url = 'https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/main/assets/languages.json'

[logging]
# level is the logging level.
# Valid values: "debug", "info", "warn", "error".
# Make sure to use debug if you're sumbitting an issue.
level = 'info'

# output is the output destination for logs.
# Valid values: "file" (logs to a file) or "stdout" (logs to the console).
output = 'file'

```

---

## Known Issues

Before opening an issue make sure to check for these known problems:
1. **cant see "show repository button":** refer to this existing issue https://github.com/zerootoad/discord-rpc-lsp/issues/3
2. **language server exiting/not working:** this is prolly means there's an error at runtime, so check ur editor logs, locate the logs from the lsp and send them with the issue.

---

## Assets

If you'd like to add custom assets (e.g., editor icons), follow these steps:

1. Add your asset (e.g., `my-editor.png`) to the `assets/icons/` directory via a pull request.
2. Wait for the repository maintaners to merge it.
   
---

## Contributing

Contributions are welcome! If you'd like to contribute, please:

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Submit a pull request.

---

## License

This project is licensed under the GNU 3.0 License. See the [LICENSE](LICENSE) file for details.

---

## Resources

- [zed-discord-presence](https://github.com/xHyroM/zed-discord-presence) for bare understanding and implementation.
- [rich-go](https://github.com/hugolgst/rich-go) for Discord Rich Presence.
- [glsp](https://github.com/tliron/glsp) lsp stuff
- [LSP Specification](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/) for the LSP implementation.

---
