package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
	ViewTimer   *time.Timer
	IdleAfter   time.Duration
	ViewAfter   time.Duration
	Client      *client.Client
	LangMaps    *client.LangMaps
	ElapsedTime *time.Time
	Config      *client.Config
	Mutex       sync.Mutex
	IsIdle      bool
	IsView      bool
}

func NewLSPHandler(name string, version string, config *client.Config) (*LSPHandler, error) {
	client.Debug("Creating new LSP handler", map[string]any{
		"name":    name,
		"version": version,
	})

	start := time.Now()
	langMaps, err := client.LoadLangMaps(config.LanguageMaps.URL)
	if err != nil {
		client.Error("Failed to load language maps", map[string]any{
			"error": err,
		})
		return nil, fmt.Errorf("failed to load language maps: %w", err)
	}
	client.WithDuration(start, "Loaded language maps", map[string]any{
		"url": config.LanguageMaps.URL,
	})

	idleAfter, err := time.ParseDuration(config.Lsp.IdleAfter)
	if err != nil {
		client.Error("Failed to parse timeout duration, using default 5m", map[string]any{
			"error": err,
		})
		idleAfter = 5 * time.Minute
	}

	viewAfter, err := time.ParseDuration(config.Lsp.ViewAfter)
	if err != nil {
		client.Error("Failed to parse timeout duration, using default 5m", map[string]any{
			"error": err,
		})
		viewAfter = 5 * time.Minute
	}
	return &LSPHandler{
		Name:      name,
		Version:   version,
		Client:    &client.Client{},
		LangMaps:  &langMaps,
		IdleAfter: idleAfter,
		ViewAfter: viewAfter,
		Config:    config,
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

	h.IdleTimer = time.AfterFunc(h.IdleAfter, func() {
		h.IsIdle = true
		h.ElapsedTime = nil

		err := client.ClearDiscordActivity(h.Config, h.Config.Discord.Activity.IdleAction, "", h.Client.WorkspaceName, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName)
		if err != nil {
			client.Error("Failed to update Discord activity", map[string]any{
				"error": err,
			})
		}
		h.ElapsedTime = nil
	})
}

func (h *LSPHandler) ResetViewTimer(filename string) {
	if h.ViewTimer != nil {
		h.ViewTimer.Stop()
	}

	h.IsView = false

	h.ViewTimer = time.AfterFunc(h.ViewAfter, func() {
		h.IsView = true
		err := client.UpdateDiscordActivity(h.Config, h.Config.Discord.Activity.ViewAction, filename, h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
		if err != nil {
			client.Error("Failed to update Discord activity", map[string]any{
				"error": err,
			})
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
	client.Info("Initializing server", map[string]any{
		"params": params,
	})

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
		client.Error("Failed to parse retry_after duration, using 1 minute", map[string]any{
			"error": err,
		})
		retryafter = 1 * time.Minute
	}
	for {
		err := client.Login(string(h.Client.ApplicationID))
		if err == nil {
			break
		}

		client.Error("Failed to create Discord RPC client, retrying in 1 minute", map[string]any{
			"error": err,
		})

		time.Sleep(retryafter)
	}

	var rootURI string
	if params.RootURI != nil {
		rootURI = string(*params.RootURI)
	} else {
		rootPath, err := os.Getwd()
		if err != nil {
			client.Error("Failed to get root path, setting to temp", map[string]any{
				"error": err,
			})

			rootURI = "file://" + os.TempDir()
		} else {
			rootURI = "file://" + rootPath
		}
	}
	h.Client.RootURI = rootURI

	client.Info("Root URI set", map[string]any{
		"rootURI": h.Client.RootURI,
	})

	if !strings.Contains(rootURI, os.TempDir()) {
		h.Client.WorkspaceName = utils.GetFileName(h.Client.RootURI)

		workspacePath := filepath.Dir(h.Client.RootURI) + "/" + utils.GetFileName(h.Client.RootURI)
		client.Info("Workspace path set", map[string]any{
			"workspacePath": workspacePath,
		})

		remoteUrl, branchName, err := client.GetGitRepositoryInfo(workspacePath)
		if err != nil {
			client.Error("Failed to get git repository info", map[string]any{
				"error": err,
			})
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
	client.Info("Initialized server", nil)
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
	client.Info("Shutdown request received", nil)
	client.Logout()

	return nil
}

func (h *LSPHandler) exit(ctx *glsp.Context) error {
	client.Info("Exit notification received", nil)
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

	client.Info("Opened file", map[string]any{
		"fileName": fileName,
		"language": h.CurrentLang,
		"params":   params,
	})

	h.ResetIdleTimer()

	go func() {
		err := client.UpdateDiscordActivity(h.Config, h.Config.Discord.Activity.ViewAction, fileName, h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
		if err != nil {
			client.Error("Failed to update Discord activity", map[string]any{
				"error": err,
			})
		}
	}()

	return nil
}

func (h *LSPHandler) didClose(ctx *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	fileName := utils.GetFileName(string(params.TextDocument.URI))

	client.Info("File closed", map[string]any{
		"fileName": fileName,
	})

	h.ResetIdleTimer()

	h.CurrentLang = ""

	go func() {
		err := client.UpdateDiscordActivity(h.Config, "No file open", "", h.Client.WorkspaceName, "", h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
		if err != nil {
			client.Error("Failed to update Discord activity", map[string]any{
				"error": err,
			})
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

	client.Info("Changed file", map[string]any{
		"fileName": fileName,
		"language": h.CurrentLang,
		"params":   params,
	})

	h.ResetIdleTimer()
	h.ResetViewTimer(fileName)

	go func() {
		if len(params.ContentChanges) > 0 {
			change := params.ContentChanges[0]

			switch change := change.(type) {
			case protocol.TextDocumentContentChangeEvent:
				var activity string
				if change.Range != nil && h.Config.Discord.Activity.EditingInfo {
					line := utils.EvalOffset(fmt.Sprintf("%d%s", change.Range.Start.Line, h.Config.Lsp.LineOffset))
					activity = h.Config.Discord.Activity.EditAction + " - In line " + strconv.Itoa(line)
				} else {
					activity = h.Config.Discord.Activity.EditAction
				}
				err := client.UpdateDiscordActivity(h.Config, activity, fileName, h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
				if err != nil {
					client.Error("Failed to update Discord activity", map[string]any{
						"error": err,
					})
				}

			case protocol.TextDocumentContentChangeEventWhole:
				err := client.UpdateDiscordActivity(h.Config, h.Config.Discord.Activity.EditAction, fileName, h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
				if err != nil {
					client.Error("Failed to update Discord activity", map[string]any{
						"error": err,
					})
				}

			default:
				err := client.UpdateDiscordActivity(h.Config, h.Config.Discord.Activity.EditAction, fileName, h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
				client.Warn("Unknown content change type", map[string]any{
					"changeType": fmt.Sprintf("%T", change),
				})
				if err != nil {
					client.Error("Failed to update Discord activity", map[string]any{
						"error": err,
					})
				}
			}
		} else {
			err := client.UpdateDiscordActivity(h.Config, h.Config.Discord.Activity.EditAction, fileName, h.Client.WorkspaceName, h.CurrentLang, h.Client.Editor, h.Client.GitRemoteURL, h.Client.GitBranchName, h.ElapsedTime)
			if err != nil {
				client.Error("Failed to update Discord activity", map[string]any{
					"error": err,
				})
			}
		}
	}()

	return nil
}
