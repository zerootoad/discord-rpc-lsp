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

- [x] LSP rewrite using glsp instead of go-lsp.
- [x] Add easy configuration. (pretty straight forward)
- [ ] Fix discord rich presence buttons not showing. (might be on discord side)
- [ ] Add diagnostics to the discord activity, best guess (zk way): [refreshDiagnosticsOfDocument](https://github.com/zk-org/zk/blob/68e6b70eaefdf8344065fcec39d5419dc80d6a02/internal/adapter/lsp/server.go#L556)

---

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
# application_id is the Discord Application ID for Rich Presence.
# This is optional, as the lsp handles it based on the editor being used.
application_id = ''

# small_usage determines what is displayed in the small icon tooltip.
# Valid values: "language" or "editor".
small_usage = 'language'

# large_usage determines what is displayed in the large icon tooltip.
# Valid values: "language" or "editor".
large_usage = 'editor'

# retry_after is the duration to wait before retrying in case it fails to create the discord rpc client.
# Must be a valid duration string (e.g., "1m", "30s").
retry_after = '1m'

[discord.activity]
# The discord activity is customizable via placeholders.
# 
# List of avaible placeholder:
# {action} : holds the action being executed (e.g., "Idling", "Editing")
# {filename} : holds the name of current file.
# {workspace} : holds the workspace name.
# {editor} : holds the editor name (e.g., "helix", "neovim")
# {language} : holds the language name of the current file.

# state is the first line of the activity status.
state = '{action} {filename}'

# details is the second line of the activity status.
details = 'In {workspace}'

# large_image is the URL for the large icon.
large_image = 'https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/main/assets/icons/{editor}.png'

# large_text is the tooltip text for the large icon.
large_text = '{editor}'

# small_image is the URL for the small icon.
small_image = 'https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/main/assets/icons/{language}.png'

# small_text is the tooltip text for the small icon.
small_text = 'Coding in {language}'

# timestamp determines whether to display a timestamp in the activity.
# If true, the time since the activity started will be shown.
timestamp = true

[lsp]
# timeout is the duration after which the LSP will enable idling if no activity is detected.
# Must be a valid duration string (e.g., "5m", "30s").
timeout = '5m'

[language_maps]
# url is the URL to a JSON file containing mappings of file extensions to programming languages.
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
