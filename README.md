# Discord Rich Presence LSP

A Language Server Protocol (LSP) to share what you're coding on Discord. This LSP integrates with your editor to display your current coding activity (file, language, and more) as a rich presence on Discord. 

![image](https://github.com/user-attachments/assets/42b6a24e-b214-473c-9de9-1708c09d0a3a)
![image](https://github.com/user-attachments/assets/20bdd2a2-fe46-45d7-a3a5-0a0326ad94f1)

---

## Features

- Displays the file you're currently editing.
- Shows the programming language you're using.
- Includes Git repository information (branch and remote URL).
- Customizable rich presence with editor-specific icons.
- Supports multiple editors via LSP.

---

## TODO

- [ ] Add easy configuration.
- [x] Idling

---

## Build Steps

### Prerequisites

- Go (version 1.21 or higher)
- Git
- Discord (with Rich Presence enabled)

### Steps

1. **Clone the Repository**

   ```bash
   git clone https://github.com/zerootoad/discord-rich-presence-lsp.git
   cd discord-rich-presence-lsp
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

For other editors, refer to their documentation on how to configure LSP servers. The process is similar: specify the path to the `discord-rich-presence-lsp` binary in your editor's LSP configuration.

---

## Configuration

Currently, configuration is done by modifying the source code. Future updates will include a configuration file for easier customization.

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
- [LSP Specification](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/) for the LSP implementation.

---
