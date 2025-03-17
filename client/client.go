package client

import (
	"github.com/go-git/go-git/v5"
)

type Client struct {
	ApplicationID string
	Editor        string
	RootURI       string
	WorkspaceName string
	GitRemoteURL  string
	GitBranchName string
}

func GetGitRepositoryInfo(workspacePath string) (remoteURL, branchName string, err error) {
	repo, err := git.PlainOpen(workspacePath)
	if err != nil {
		return "", "", err
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return "", "", err
	}
	remoteURL = remote.Config().URLs[0]

	head, err := repo.Head()
	if err != nil {
		return "", "", err
	}
	branchName = head.Name().Short()

	return remoteURL, branchName, nil
}
