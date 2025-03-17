package handler

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"

	"github.com/zerootoad/discord-rpc-lsp/client"
)

type LSPHandler struct {
	Shutdown    bool
	CurrentLang string
	IdleTimer   *time.Timer
	Client      *client.Client
}

func NewLSPHandler() *LSPHandler {
	return &LSPHandler{
        
		Client: &client.Client{},
	}
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
			h.Client.RootURI = string(params.RootURI)
			h.Client.WorkspaceName = client.GetFileName(h.Client.RootURI)

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
			log.Println("Exit notification received")
			if h.Shutdown {
				os.Exit(0)
			} else {
				os.Exit(1)
			}

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

			fileName := client.GetFileName(string(params.TextDocument.URI))
			h.CurrentLang = params.TextDocument.LanguageID

			log.Printf("File opened: %s (Language: %s)", fileName, h.CurrentLang)

			go func() {
				client.UpdateDiscordActivity("Watching", "Editing "+fileName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName)
			}()

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

			fileName := client.GetFileName(string(params.TextDocument.URI))

			log.Printf("File changed: %s (Language: %s)", fileName, h.CurrentLang)

			if len(params.ContentChanges) > 0 {
				changes := params.ContentChanges[0]
				if changes.Range != nil {
					log.Printf("File changed from {%v}:{%v}", changes.Range.Start, changes.Range.End)
				} else {
					log.Printf("File changed (no range information)")
				}
				log.Printf("Change text: %s", changes.Text)
			} else {
				log.Printf("File changed (no content changes)")
			}

			go func() {
				client.UpdateDiscordActivity("Writing", "Editing "+fileName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName)
			}()

		default:
			log.Printf("Unhandled method: %s", req.Method)
		}
	}()
}
