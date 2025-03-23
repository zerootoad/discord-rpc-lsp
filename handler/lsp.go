package handler

import (
	"fmt"
	log "github.com/sirupsen/logrus"
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
	log.WithFields(log.Fields{
		"name":    name,
		"version": version,
	}).Info("Creating new LSP handler")

	langMaps, err := client.LoadLangMaps("https://raw.githubusercontent.com/zerootoad/discord-rich-presence-lsp/refs/heads/main/assets/languages.json")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to load language maps")
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
		err := client.ClearDiscordActivity("In "+h.Client.WorkspaceName, "Idling", h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to update Discord activity")
		}
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
	log.WithFields(log.Fields{
		"params": params,
	}).Info("Initializing server")

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

		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to create Discord RPC client, retrying in 1 minute")

		time.Sleep(1 * time.Minute)
	}

	h.Client.RootURI = string(*params.RootURI)
	log.WithFields(log.Fields{
		"rootURI": h.Client.RootURI,
	}).Info("Root URI set")

	h.Client.WorkspaceName = utils.GetFileName(h.Client.RootURI)

	workspacePath := filepath.Dir(h.Client.RootURI) + "/" + utils.GetFileName(h.Client.RootURI)
	log.WithFields(log.Fields{
		"workspacePath": workspacePath,
	}).Info("Workspace path set")

	remoteUrl, branchName, err := client.GetGitRepositoryInfo(workspacePath)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to get git repository info")
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
	log.Info("Initialized server")
	return nil
}

func (h *LSPHandler) setTrace(ctx *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}

func (h *LSPHandler) shutdown(ctx *glsp.Context) error {
	h.Shutdown = true
	log.Info("Shutdown request received")
	client.Logout()

	return nil
}

func (h *LSPHandler) exit(ctx *glsp.Context) error {
	log.Info("Exit notification received")
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

	log.WithFields(log.Fields{
		"fileName": fileName,
		"language": h.CurrentLang,
		"params":   params,
	}).Info("Opened file")

	h.ResetIdleTimer()

	go func() {
		err := client.UpdateDiscordActivity("Watching "+fileName, "In "+h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to update Discord activity")
		}
	}()

	return nil
}

func (h *LSPHandler) didClose(ctx *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	fileName := utils.GetFileName(string(params.TextDocument.URI))

	log.WithFields(log.Fields{
		"fileName": fileName,
	}).Info("File closed")

	h.ResetIdleTimer()

	if h.CurrentLang != "" {
		h.CurrentLang = ""
	}

	go func() {
		err := client.UpdateDiscordActivity("No file open", "In "+h.Client.WorkspaceName, "", h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to update Discord activity")
		}
	}()

	return nil
}

func (h *LSPHandler) didChange(ctx *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	fileName := utils.GetFileName(string(params.TextDocument.URI))
	h.CurrentLang = h.LangMaps.GetLanguage(fileName)
	if h.CurrentLang == "" {
		h.CurrentLang = "text"
	}

	log.WithFields(log.Fields{
		"fileName": fileName,
		"language": h.CurrentLang,
		"params":   params,
	}).Info("Changed file")

	h.ResetIdleTimer()

	go func() {
		if len(params.ContentChanges) > 0 {
			change := params.ContentChanges[0]

			switch change := change.(type) {
			case protocol.TextDocumentContentChangeEvent:
				var activity string
				if change.Range != nil {
					line := change.Range.Start.Line
					activity = "In line " + strconv.Itoa(int(line))
				} else {
					activity = "Editing " + fileName
				}
				err := client.UpdateDiscordActivity(activity, "In "+h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("Failed to update Discord activity")
				}

			case protocol.TextDocumentContentChangeEventWhole:
				err := client.UpdateDiscordActivity("Editing "+fileName, "In "+h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("Failed to update Discord activity")
				}

			default:
				log.WithFields(log.Fields{
					"changeType": fmt.Sprintf("%T", change),
				}).Warn("Unknown content change type")
				err := client.UpdateDiscordActivity("Editing "+fileName, "In "+h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("Failed to update Discord activity")
				}
			}
		} else {
			err := client.UpdateDiscordActivity("Editing "+fileName, "In "+h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Failed to update Discord activity")
			}
		}
	}()

	return nil
}
