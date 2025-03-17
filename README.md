# Discord Rich Presence LSP

A Language Server Protocol (LSP) to share what you're coding on Discord. This LSP integrates with your editor to display your current coding activity (file, language, and more) as a rich presence on Discord. 

![image](https://github.com/user-attachments/assets/383ba71d-c492-497f-a8c5-553d0363c24e)
![image](https://github.com/user-attachments/assets/da7e4efb-d8c0-4d16-9817-ec2f0964a649)


---

## Features

- Displays the file you're currently editing.
- Shows the programming language you're using.
- Includes Git repository information (branch and remote URL).
- Customizable rich presence with editor-specific icons.
- Supports multiple editors via LSP.

---

## TODO

- [x] Fix issues with opening and closing files that cause the LSP to stop detecting changes. *(Possibly fixed)*
- [x] Improve the rich presence embed. *(Mostly improved; needs additional details like line/column, diagnostics, etc.)*
- [x] Add assets. *(Make a pull request or open an issue if you need an asset added.)*
- [x] Improve code structure. *(Should be good enough for now.)*
- [x] Implement structs to make activity handling easier.
- [x] Fix the rich presence title. *(Turns out it was the application ID name; looking to implement a way to set it based on the editor being used.)*
- [ ] Add easy configuration options.
- [ ] Implement Idling

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

### **Visual Studio Code (VS Code)**

1. Install the [LSP Extension](https://marketplace.visualstudio.com/items?itemName=matklad.lsp) if you don't already have it.
2. Open your VS Code settings (`Ctrl + ,` or `Cmd + ,`).
3. Add the following configuration to your `settings.json`:

   ```json
   soon..
   ```

4. Replace `path/to/discord-rpc-lsp` with the actual path to the built binary.

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
4. Add '"discord-rpc"' for the choosen languages:
   ```toml
   [[language]]
   name = "go"
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

## Contributors
**@zerootoad**
**@nyrilol**
