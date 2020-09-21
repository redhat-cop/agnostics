package v1

import (
	"github.com/go-git/go-git/v5"
	"time"
	"errors"
)

// ErrGitRepoNoURLs error when the Repository doesn't have any remote URLs defined
var ErrGitRepoNoURLs = errors.New("Git repository has no remote URLs")

// NewGitCommit constructor for GitCommit
func NewGitCommit(configRepo *git.Repository) (GitCommit, error) {
	remote, err := configRepo.Remote("origin")
	if err != nil {
		return GitCommit{}, err
	}

	if len(remote.Config().URLs) == 0 {
		return GitCommit{}, ErrGitRepoNoURLs
	}
	origin := remote.Config().URLs[0]

	if rev, err := configRepo.ResolveRevision("HEAD") ; err == nil {
		if head, err := configRepo.CommitObject(*rev) ; err == nil {
			return GitCommit{
				Hash: head.Hash.String(),
				Author: head.Author.Name,
				Date: head.Author.When.UTC().Format(time.RFC3339),
				Origin: origin,
			}, nil
		}
		return GitCommit{}, err
	}
	return GitCommit{}, err
}
