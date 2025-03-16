package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/altfoxie/drpc"        // discord rich presence
	"github.com/sourcegraph/go-lsp"   // lsp sh
	"github.com/sourcegraph/jsonrpc2" // lsp sh
)

// LSP Handler bs
//
// Resources:
// - https://github.com/AzimovParviz/Discord-rich-presence-LSP/
// - https://github.com/xHyroM/zed-discord-presence
// - https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/ for textDocuments

type LSPHandler struct {
	shutdown bool
	client   *drpc.Client
}

func (h *LSPHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	go func() {
		// log.Printf("Received request: %s", req.Method)

		// if req.Params != nil {
		// log.Printf("Request params: %s", string(*req.Params))
		// } else {
		// log.Printf("Request params: nil")
		// }

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
			h.shutdown = true
			log.Println("Shutdown request received")
			if err := conn.Reply(ctx, req.ID, nil); err != nil {
				log.Printf("Failed to reply to shutdown request: %v", err)
			}

		case "exit":
			log.Println("Exit notification received")
			if h.shutdown {
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

			log.Printf("File opened: %s", params.TextDocument.URI)
			log.Printf("File language: %s", params.TextDocument.LanguageID)
			log.Printf("File version: %d", params.TextDocument.Version)

			go func() {
				err := h.client.SetActivity(drpc.Activity{
					Details: "Editing " + string(params.TextDocument.URI),
					State:   "Language: " + string(params.TextDocument.LanguageID),
					Timestamps: &drpc.Timestamps{
						Start: time.Now(),
					},
					Assets: &drpc.Assets{
						LargeImage: "code",
						LargeText:  "Coding in " + string(params.TextDocument.LanguageID),
					},
				})
				if err != nil {
					log.Printf("Failed to update Rich Presence: %v", err)
				}
			}()

		case "textDocument/didClose":
			if req.Params == nil {
				log.Printf("Error: didClose request has nil params")
				return
			}

			var params lsp.DidCloseTextDocumentParams
			if err := json.Unmarshal(*req.Params, &params); err != nil {
				log.Printf("Error unmarshaling didClose params: %v", err)
				return
			}

			log.Printf("File closed: %s", params.TextDocument.URI)
			go func() {
				err := h.client.SetActivity(drpc.Activity{
					Details: "No file open",
					State:   "Idle",
					Timestamps: &drpc.Timestamps{
						Start: time.Now(),
					},
					Assets: &drpc.Assets{
						LargeImage: "code",
						LargeText:  "Idle",
					},
				})
				if err != nil {
					log.Printf("Failed to update Rich Presence: %v", err)
				}
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

			log.Printf("File changed: %s", params.TextDocument.URI)
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
				err := h.client.SetActivity(drpc.Activity{
					Details: "Editing " + string(params.TextDocument.URI),
					State:   "Changes made",
					Timestamps: &drpc.Timestamps{
						Start: time.Now(),
					},
					Assets: &drpc.Assets{
						LargeImage: "code",
						LargeText:  "Coding",
					},
				})
				if err != nil {
					log.Printf("Failed to update Rich Presence: %v", err)
				}
			}()

		default:
			log.Printf("Unhandled method: %s", req.Method)
		}
	}()
}

func main() {
	logFile, err := os.OpenFile("discord-lsp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	client, err := drpc.New("975346661540909056") // application id
	if err != nil {
		log.Fatalf("Failed to create Discord RPC client: %v", err)
	}
	defer client.Close()

	handler := &LSPHandler{
		client: client,
	}
	stream := jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{})
	conn := jsonrpc2.NewConn(context.Background(), stream, handler)

	log.Println("LSP server started!")
	<-conn.DisconnectNotify()
	log.Println("LSP server stopped!")
}

// stdin n stdout bs
type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (stdrwc) Close() error {
	return nil
}
