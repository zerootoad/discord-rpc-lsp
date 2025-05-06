package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

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
	ViewTimer   *time.Timer
	Timeout     time.Duration
	Client      *client.Client
	LangMaps    *client.LangMaps
	ElapsedTime *time.Time
	Config      *client.Config
	Mutex       sync.Mutex
	IsIdle      bool
	IsView      bool
}

func NewLSPHandler(name string, version string, config *client.Config) (*LSPHandler, error) {
	log.WithFields(log.Fields{
		"name":    name,
		"version": version,
	}).Info("Creating new LSP handler")

	langMaps, err := client.LoadLangMaps(config.LanguageMaps.URL)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to load language maps")
		return nil, fmt.Errorf("failed to load language maps: %w", err)
	}
	timeout, err := time.ParseDuration(config.Lsp.Timeout)
	if err != nil {
		log.Errorf("Failed to parse timeout duration: %v", err)
		timeout = 5 * time.Minute
	}
	return &LSPHandler{
		Name:     name,
		Version:  version,
		Client:   &client.Client{},
		LangMaps: &langMaps,
		Timeout:  timeout,
		Config:   config,
	}, nil
}

func (h *LSPHandler) ResetIdleTimer() {
	if h.IdleTimer != nil {
		h.IdleTimer.Stop()
	}

	if h.ElapsedTime == nil {
		now := time.Now()
		h.ElapsedTime = &now
		h.IsIdle = false
	}

	h.IdleTimer = time.AfterFunc(h.Timeout, func() {
		h.IsIdle = true
		h.ElapsedTime = nil

		err := client.ClearDiscordActivity(h.Config, h.Config.Discord.Activity.IdleAction, "", h.Client.WorkspaceName, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to update Discord activity")
		}
		h.ElapsedTime = nil
	})
}

func (h *LSPHandler) ResetViewTimer(filename string) {
	if h.ViewTimer != nil {
		h.ViewTimer.Stop()
	}

	h.IsView = false

	h.ViewTimer = time.AfterFunc(1*time.Minute, func() {
		h.IsView = true
		err := client.UpdateDiscordActivity(h.Config, h.Config.Discord.Activity.ViewAction, filename, h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to update Discord activity")
		}
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

	if params == nil {
		return nil, fmt.Errorf("initialize params cannot be nil")
	}

	capabilities := h.Handler.CreateServerCapabilities()

	h.Client.Editor = strings.ToLower(params.ClientInfo.Name)
	h.Client.ApplicationID = ""
	if h.Config.Discord.ApplicationID != "" {
		h.Client.ApplicationID = h.Config.Discord.ApplicationID
	} else {
		switch h.Client.Editor {
		case "neovim":
			h.Client.ApplicationID = "1352048301633044521" // Neovim
		case "helix":
			h.Client.ApplicationID = "1351256971059396679" // Helix
		default:
			h.Client.ApplicationID = "1351257618227920896" // Code
		}
	}

	retryafter, err := time.ParseDuration(h.Config.Discord.RetryAfter)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to parse retry_after duration using 1 minute")
		retryafter = 1 * time.Minute
	}
	for {
		err := client.Login(string(h.Client.ApplicationID))
		if err == nil {
			break
		}

		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to create Discord RPC client, retrying in 1 minute")

		time.Sleep(retryafter)
	}

	var rootURI string
	if params.RootURI != nil {
		rootURI = string(*params.RootURI)
	} else {
		rootPath, err := os.Getwd()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to get root path, setting to temp")

			rootURI = "file://" + os.TempDir()
		} else {
			rootURI = "file://" + rootPath
		}
	}
	h.Client.RootURI = rootURI

	log.WithFields(log.Fields{
		"rootURI": h.Client.RootURI,
	}).Info("Root URI set")

	if !strings.Contains(rootURI, os.TempDir()) {
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
	} else {
		h.Client.WorkspaceName = h.Client.Editor
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
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

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
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

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
		err := client.UpdateDiscordActivity(h.Config, h.Config.Discord.Activity.ViewAction, fileName, h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to update Discord activity")
		}
	}()

	return nil
}

func (h *LSPHandler) didClose(ctx *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	fileName := utils.GetFileName(string(params.TextDocument.URI))

	log.WithFields(log.Fields{
		"fileName": fileName,
	}).Info("File closed")

	h.ResetIdleTimer()

	if h.CurrentLang != "" {
		h.CurrentLang = ""
	}

	go func() {
		err := client.UpdateDiscordActivity(h.Config, "No file open", "", h.Client.WorkspaceName, "", h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to update Discord activity")
		}
	}()

	return nil
}

func (h *LSPHandler) didChange(ctx *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

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
	h.ResetViewTimer(fileName)

	go func() {
		if len(params.ContentChanges) > 0 {
			change := params.ContentChanges[0]

			switch change := change.(type) {
			case protocol.TextDocumentContentChangeEvent:
				var activity string
				if change.Range != nil && h.Config.Discord.Activity.EditingInfo {
					line := change.Range.Start.Line
					activity = h.Config.Discord.Activity.EditAction + " - In line " + strconv.Itoa(int(line))
				} else {
					activity = h.Config.Discord.Activity.EditAction
				}
				err := client.UpdateDiscordActivity(h.Config, activity, fileName, h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("Failed to update Discord activity")
				}

			case protocol.TextDocumentContentChangeEventWhole:
				err := client.UpdateDiscordActivity(h.Config, h.Config.Discord.Activity.EditAction, fileName, h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("Failed to update Discord activity")
				}

			default:
				err := client.UpdateDiscordActivity(h.Config, h.Config.Discord.Activity.EditAction, fileName, h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
				log.WithFields(log.Fields{
					"changeType": fmt.Sprintf("%T", change),
				}).Warn("Unknown content change type")
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("Failed to update Discord activity")
				}
			}
		} else {
			err := client.UpdateDiscordActivity(h.Config, h.Config.Discord.Activity.EditAction, fileName, h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Failed to update Discord activity")
			}
		}
	}()

	return nil
}
