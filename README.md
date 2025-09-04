# Discord Rich Presence LSP

A Language Server Protocol (LSP) to share what you're coding on Discord. This LSP integrates with your editor to display your current coding activity (file, language, Git info, etc.) as a rich presence on Discord.

![image](https://github.com/user-attachments/assets/3089b1ab-0f04-46d3-ae59-ed6207a853f4)
![image](https://github.com/user-attachments/assets/ddbd5f14-65f2-4ab0-a0c5-0b5eb3b36795)

---

## Features

* Displays the file you're currently editing.
* Shows the programming language you're using.
* Includes Git repository information (branch and remote URL).
* Customizable rich presence with editor-specific icons.
* Supports multiple editors via LSP.

---

## TODO

* [x] Implement zerolog for logging.
* [x] Create tagged releases on GitHub (e.g., v1.0.0).
* [ ] Improve project code.
* [ ] Improve customization options.

---

## Installation

### Prerequisites

* Go 1.21+
* Git
* Discord (with Rich Presence enabled)

---

### Linux

**AUR (Arch/Manjaro)**

```bash
yay -S discord-rpc-lsp-git
```

**Manual via Go**

```bash
go install github.com/zerootoad/discord-rpc-lsp@latest
```

Binary installs to `$(go env GOPATH)/bin` (usually `~/go/bin`). Add to PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

---

### macOS

**Manual via Go**

```bash
go install github.com/zerootoad/discord-rpc-lsp@latest
```

Binary installs to `$(go env GOPATH)/bin` (usually `~/go/bin`). Add to PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

---

### Windows

**Manual via Go**

```powershell
go install github.com/zerootoad/discord-rpc-lsp@latest
```

Binary installs to `%USERPROFILE%\go\bin`. Add this to your **Environment Variables → PATH**.

---

### Build from Source (Optional)

```bash
git clone https://github.com/zerootoad/discord-rpc-lsp.git
cd discord-rpc-lsp
go mod tidy
go build
```

Run the server:

```bash
# Linux/macOS
./discord-rpc-lsp

# Windows
discord-rpc-lsp.exe
```

---

## Adding to Editors

### Supported Editors

This LSP works with any editor that supports the Language Server Protocol (LSP).

---

### Neovim

1. Install `nvim-lspconfig`.
2. Add to your `init.lua`:

```lua
local lspconfig = require('lspconfig')
local configs = require('lspconfig.configs')

configs.discord_rpc = {
    default_config = {
        cmd = { "discord-rpc-lsp" },  -- full path if not in PATH
        filetypes = {"*"},             -- adjust to your languages
        root_dir = function(fname)
            return lspconfig.util.root_pattern('.git')(fname) or vim.fn.getcwd()
        end,
        settings = {},
    },
}
```

---

### Helix

`~/.config/helix/languages.toml`:

```toml
[language-server.discord-rpc]
command = "discord-rpc-lsp"  # full path if needed

[[language]]
name = "go"
language-servers = ["discord-rpc"]

[[language]]
name = "python"
language-servers = ["discord-rpc"]
```

Add additional languages as needed.

---

### Visual Studio Code

`settings.json`:

```json
{
  "languageserver": {
    "discord-rpc-lsp": {
      "command": "discord-rpc-lsp",
      "args": [],
      "filetypes": ["go","python","javascript","typescript","rust","lua","c","cpp","java","text"],
      "rootPatterns": [".git","."],
      "trace.server": "verbose"
    }
  }
}
```

---

### Zed Editor

Zed automatically detects LSP servers:

1. Ensure `discord-rpc-lsp` is in your PATH.
2. Open Zed and a project/file → Discord Rich Presence updates automatically.

Optional wrapper script:

```bash
#!/bin/bash
export PATH=$PATH:$(go env GOPATH)/bin
zed "$@"
```

---

## Configuration

Configuration is done via `config.toml` in the configuration directory:

* **Linux/macOS:** `~/.discord-rpc-lsp/`
* **Windows:** `%APPDATA%\Roaming\.discord-rpc-lsp\`

Default configuration includes:

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
idle_after = '5m'
# The duration after which the editing mode will go in viewing if no changes are applied.
# Must be a valid duration string (e.g., "5m", "30s").
view_after = '30s'
# This indicates how much u should offset the line for, this is a fix incase ur line index isnt right.
# Must be a valid expression ("+1", "+ 2", "-3", "- 4").
# If ur line is off by one, change this to "+0" or "-0".
line_offset = '+1'


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

1. “Show repository” button may not appear: check [issue #3](https://github.com/zerootoad/discord-rpc-lsp/issues/3).
2. LSP exiting/not working: check editor logs and LSP output logs for runtime errors and open an issue.

---

## Assets

To add custom assets (icons, etc.):

1. Add your asset to `assets/icons/`.
2. Open a pull request.
3. Wait for merge.

---

## Contributing

1. Fork the repository.
2. Create a branch for your feature or bugfix.
3. Submit a pull request.

---

## License

GNU 3.0 License — see [LICENSE](LICENSE).

---

## Resources

* [zed-discord-presence](https://github.com/xHyroM/zed-discord-presence)
* [rich-go](https://github.com/hugolgst/rich-go)
* [glsp](https://github.com/tliron/glsp)
* [LSP Specification](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/)

---
