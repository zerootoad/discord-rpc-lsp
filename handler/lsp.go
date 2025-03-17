package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/zerootoad/discord-rpc-lsp/client"
	"github.com/zerootoad/discord-rpc-lsp/utils"
)

type LSPHandler struct {
	Shutdown      bool
	CurrentLang   string
	IdleTimer     *time.Timer
	Client        *client.Client
	LangMaps      *client.LangMaps
	ProblemsCount int
}

func NewLSPHandler() (*LSPHandler, error) {
	langMaps, err := client.LoadLangMaps("https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/languages.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load language maps: %w", err)
	}

	return &LSPHandler{
		Client:   &client.Client{},
		LangMaps: &langMaps,
	}, nil
}

func (h *LSPHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	go func() {
		log.Printf("Received request: %s", req.Method)
		log.Printf("Request params: %s", req.Params)

		switch req.Method {
		case "initialize":
			if req.Params == nil {
				log.Printf("Error: initialize request has nil params")
				return
			}

			var params lsp.InitializeParams
			if err := json.Unmarshal(*req.Params, &params); err != nil {
				log.Printf("Error unmarshaling initialize params: %v", err)
				return
			}

			h.Client.Editor = params.ClientInfo.Name
			h.Client.ApplicationID = ""
			switch h.Client.Editor {
			case "Neovim":
				h.Client.ApplicationID = "1351256847612514390" // Nvim
			case "helix":
				h.Client.ApplicationID = "1351256971059396679" // Helix
			default:
				h.Client.ApplicationID = "1351257618227920896" // Code
			}

			for {
				err := client.Login(string(h.Client.ApplicationID))
				if err == nil {
					break
				}

				log.Fatalf("Failed to create Discord RPC client: %v (retrying in 1 minute)", err)

				time.Sleep(1 * time.Minute)
			}

			h.Client.RootURI = string(params.RootURI)
			h.Client.WorkspaceName = utils.GetFileName(h.Client.RootURI)

			workspacePath := filepath.Dir(h.Client.RootURI)
			remoteUrl, branchName, err := client.GetGitRepositoryInfo(workspacePath)
			if err != nil {
				log.Printf("Failed to get git repository info: %v", err)
			} else {
				h.Client.GitRemoteURL = remoteUrl
				h.Client.GitBranchName = branchName
			}

			response := lsp.InitializeResult{
				Capabilities: lsp.ServerCapabilities{
					TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
						Options: &lsp.TextDocumentSyncOptions{
							OpenClose: true,
							Change:    lsp.TDSKIncremental,
						},
					},
				},
			}

			if err := conn.Reply(ctx, req.ID, response); err != nil {
				log.Printf("Failed to reply to initialize request: %v", err)
			} else {
				log.Printf("Initialized LSP server with capabilities: %+v", response.Capabilities)
			}

		case "initialized":
			log.Println("Client initialized")

		case "shutdown":
			h.Shutdown = true
			log.Println("Shutdown request received")
			if err := conn.Reply(ctx, req.ID, nil); err != nil {
				log.Printf("Failed to reply to shutdown request: %v", err)
			}
			client.Logout()

		case "exit":
			h.exit()

		case "textDocument/didOpen":
			if req.Params == nil {
				log.Printf("Error: didOpen request has nil params")
				return
			}

			var params lsp.DidOpenTextDocumentParams
			if err := json.Unmarshal(*req.Params, &params); err != nil {
				log.Printf("Error unmarshaling didOpen params: %v", err)
				return
			}

			h.didOpen(params)

		case "textDocument/didChange":
			if req.Params == nil {
				log.Printf("Error: didChange request has nil params")
				return
			}

			var params lsp.DidChangeTextDocumentParams
			if err := json.Unmarshal(*req.Params, &params); err != nil {
				log.Printf("Error unmarshaling didChange params: %v", err)
				return
			}

			h.didChange(params)

		default:
			log.Printf("Unhandled method: %s", req.Method)
		}
	}()
}

func (h *LSPHandler) didOpen(params lsp.DidOpenTextDocumentParams) {
	fileName := utils.GetFileName(string(params.TextDocument.URI))

	h.CurrentLang = h.LangMaps.GetLanguage(fileName)
	if h.CurrentLang == "" {
		h.CurrentLang = params.TextDocument.LanguageID
	}

	log.Printf("File opened: %s (Language: %s)", fileName, h.CurrentLang)

	go func() {
		client.UpdateDiscordActivity("Watching "+fileName, "In "+h.Client.Editor, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName)
	}()
}

func (h *LSPHandler) didChange(params lsp.DidChangeTextDocumentParams) {
	fileName := utils.GetFileName(string(params.TextDocument.URI))

	h.CurrentLang = h.LangMaps.GetLanguage(fileName)
	if h.CurrentLang == "" {
		h.CurrentLang = "text"
	}
	log.Printf("File changed: %s (Language: %s)", fileName, h.CurrentLang)

	go func() {
		if len(params.ContentChanges) > 0 {
			changes := params.ContentChanges[0]
			if changes.Range != nil {
				client.UpdateDiscordActivity("In line "+strconv.Itoa(changes.Range.Start.Line), "Editing "+fileName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName)
			}
		} else {
			client.UpdateDiscordActivity("Editing "+fileName, "In "+h.Client.Editor, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName)
		}
	}()
}

func (h *LSPHandler) exit() {
	log.Println("Exit notification received")
	if h.Shutdown {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
