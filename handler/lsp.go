package handler

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
	"github.com/zerootoad/discord-rpc-lsp/client"
	"github.com/zerootoad/discord-rpc-lsp/utils"
)

type LSPHandler struct {
	Name        string
	Version     string
	Handler     *protocol.Handler
	Shutdown    bool
	CurrentLang string
	IdleTimer   *time.Timer
	Timeout     time.Duration
	Client      *client.Client
	LangMaps    *client.LangMaps
	ElapsedTime *time.Time
}

func NewLSPHandler(name string, version string) (*LSPHandler, error) {
	langMaps, err := client.LoadLangMaps("https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/languages.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load language maps: %w", err)
	}
	return &LSPHandler{
		Name:     name,
		Version:  version,
		Client:   &client.Client{},
		LangMaps: &langMaps,
		Timeout:  5 * time.Minute,
	}, nil
}

func (h *LSPHandler) ResetIdleTimer() {
	if h.IdleTimer != nil {
		h.IdleTimer.Stop()
	}

	if h.ElapsedTime == nil {
		now := time.Now()
		h.ElapsedTime = &now
	}

	h.IdleTimer = time.AfterFunc(h.Timeout, func() {
		client.ClearDiscordActivity("In "+h.Client.WorkspaceName, "Idling", h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName)
		h.ElapsedTime = nil
	})
}

func (h *LSPHandler) NewServer() *server.Server {
	h.Handler = &protocol.Handler{
		Initialize:  h.initialize,
		Initialized: h.initialized,
		Shutdown:    h.shutdown,
		Exit:        h.exit,
		SetTrace:    h.setTrace,

		// textDoc notis
		TextDocumentDidOpen:   h.didOpen,
		TextDocumentDidChange: h.didChange,
		TextDocumentDidClose:  h.didClose,
	}

	return server.NewServer(h.Handler, h.Name, false)
}

func (h *LSPHandler) initialize(ctx *glsp.Context, params *protocol.InitializeParams) (any, error) {
	log.Printf("Initializing server, params: %v", params)
	capabilities := h.Handler.CreateServerCapabilities()

	h.Client.Editor = strings.ToLower(params.ClientInfo.Name)
	h.Client.ApplicationID = ""
	switch h.Client.Editor {
	case "neovim":
		h.Client.ApplicationID = "1352048301633044521" // Neovim
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

	h.Client.RootURI = string(*params.RootURI)
	log.Println(string(h.Client.RootURI))
	h.Client.WorkspaceName = utils.GetFileName(h.Client.RootURI)

	workspacePath := filepath.Dir(h.Client.RootURI) + "/" + utils.GetFileName(h.Client.RootURI)
	log.Println(string(workspacePath))
	remoteUrl, branchName, err := client.GetGitRepositoryInfo(workspacePath)
	if err != nil {
		log.Printf("Failed to get git repository info: %v", err)
	} else {
		h.Client.GitRemoteURL = remoteUrl
		h.Client.GitBranchName = branchName
	}

	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    h.Name,
			Version: &h.Version,
		},
	}, nil
}

func (h *LSPHandler) initialized(ctx *glsp.Context, params *protocol.InitializedParams) error {
	log.Println("Initialized server")
	return nil
}

func (h *LSPHandler) setTrace(ctx *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}

func (h *LSPHandler) shutdown(ctx *glsp.Context) error {
	h.Shutdown = true
	log.Println("Shutdown request received")
	client.Logout()

	return nil
}

func (h *LSPHandler) exit(ctx *glsp.Context) error {
	log.Println("Exit notification received")
	if h.IdleTimer != nil {
		h.IdleTimer.Stop()
	}
	if h.Shutdown {
		os.Exit(0)
	} else {
		os.Exit(1)
	}

	return nil
}

func (h *LSPHandler) didOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	fileName := utils.GetFileName(string(params.TextDocument.URI))

	h.CurrentLang = h.LangMaps.GetLanguage(fileName)
	if h.CurrentLang == "" {
		h.CurrentLang = params.TextDocument.LanguageID
	}
	log.Printf("File opened: %s (Language: %s)", fileName, h.CurrentLang)

	h.ResetIdleTimer()

	go func() {
		client.UpdateDiscordActivity("Watching "+fileName, "In "+h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
	}()

	return nil
}

func (h *LSPHandler) didClose(ctx *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	fileName := utils.GetFileName(string(params.TextDocument.URI))

	log.Printf("File closed: %s", fileName)

	h.ResetIdleTimer()

	if h.CurrentLang != "" {
		h.CurrentLang = ""
	}

	go func() {
		client.UpdateDiscordActivity("No file open", "In "+h.Client.WorkspaceName, "", h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
	}()

	return nil
}

func (h *LSPHandler) didChange(ctx *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	fileName := utils.GetFileName(string(params.TextDocument.URI))

	h.CurrentLang = h.LangMaps.GetLanguage(fileName)
	if h.CurrentLang == "" {
		h.CurrentLang = "text"
	}
	log.Printf("File changed: %s (Language: %s)", fileName, h.CurrentLang)

	h.ResetIdleTimer()

	go func() {
		if len(params.ContentChanges) > 0 {
			change := params.ContentChanges[0]

			switch change := change.(type) {
			case protocol.TextDocumentContentChangeEvent:
				if change.Range != nil {
					line := change.Range.Start.Line
					client.UpdateDiscordActivity("In line "+strconv.Itoa(int(line)), "Editing "+fileName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
				} else {
					client.UpdateDiscordActivity("Editing "+fileName, "In "+h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
				}

			case protocol.TextDocumentContentChangeEventWhole:
				client.UpdateDiscordActivity("Editing "+fileName, "In "+h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)

			default:
				log.Printf("Unknown content change type: %T", change)
				client.UpdateDiscordActivity("Editing "+fileName, "In "+h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
			}
		} else {
			client.UpdateDiscordActivity("Editing "+fileName, "In "+h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
		}
	}()

	return nil
}
