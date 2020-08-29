package git

import (
	//"fmt"
	"path/filepath"
	"github.com/go-git/go-git/v5"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	gitobject "github.com/go-git/go-git/v5/plumbing/object"

	"github.com/redhat-gpe/agnostics/internal/log"
	"strings"
	"io/ioutil"
	"os"
	"time"
	"errors"
)

var configRepo *git.Repository
var configRepoDir string
var configRepoCloneOptions *git.CloneOptions
var configRepoPullOptions *git.PullOptions

var ErrUpdatedTooRecently = errors.New("updated too recently")

// This function returns the current repository path as a string.
func GetRepoDir() string {
	return configRepoDir
}

// This function returns (*object.Commit, error). The Git Commit Object is the result of resolving HEAD on the config repository used by the scheduler..
// see https://pkg.go.dev/github.com/go-git/go-git/v5@v5.1.0/plumbing/object?tab=doc#Commit
// Error is nil if OK.
func GetRepoHeadCommit() (*gitobject.Commit, error) {
	if rev, err := configRepo.ResolveRevision("HEAD") ; err == nil {
		if head, err := configRepo.CommitObject(*rev) ; err == nil {
			return head, nil
		} else {
			log.Err.Println(err)
			return nil, err
		}
	} else {
		log.Err.Println(err)
		return nil, err
	}
}

func CloneRepository(url string, sshPrivateKey string) {
	// Tempdir to clone the repository
	dir, err := ioutil.TempDir("", "scheduler-config-")
	if err != nil {
		log.Err.Fatal(err)
	}

	log.Debug.Println("temporary directory for cloning", dir)

	if url[0:4] != "http" {
		log.Debug.Println("Assume SSH is used in the git URL")

		// Setup auth

		if sshPrivateKey == "" {
			sshPrivateKey = filepath.Join(os.Getenv("HOME") + "/.ssh/id_rsa")
		}
		log.Out.Println("Cloning (SSH) using private key", sshPrivateKey)

		ss := strings.FieldsFunc(url, func(r rune) bool {
			if r == '@' {
				return true
			}
			return false
		})
		auth, err := gitssh.NewPublicKeysFromFile(ss[0], sshPrivateKey, "")
		if err != nil {
			log.Err.Fatal(err)
		}
		// Clones the repository into the given dir, just as a normal git clone does
		configRepoCloneOptions = &git.CloneOptions{
			URL: url,
			SingleBranch: true,
			Auth: auth,
		}
		configRepoPullOptions = &git.PullOptions{
			SingleBranch: true,
			Auth: auth,
		}
		configRepo, err = git.PlainClone(dir, false, configRepoCloneOptions)
		if err != nil {
			log.Err.Fatal(err)
		}
	} else {
		// Clones the repository into the given dir, just as a normal git clone does
		configRepoCloneOptions = &git.CloneOptions{
			URL: url,
			SingleBranch: true,
		}
		configRepoPullOptions = &git.PullOptions{
			SingleBranch: true,
		}
		configRepo, err = git.PlainClone(dir, false, configRepoCloneOptions)
		if err != nil {
			log.Err.Fatal(err)
		}
	}
	configRepoDir = dir
}

var lastUpdated time.Time

// This function refreshes the git Worktree containing the configuration of the scheduler.
// It's basically running the equivalent of 'git pull'.
func RefreshRepository() error {
	// Do not spam git pull. Allow only one pull every 10 seconds for this process.
	if time.Now().Sub(lastUpdated) < 10 * time.Second {
		log.Debug.Println("Git repo updated recently. Ignoring.")
		return ErrUpdatedTooRecently
	}
	wt, err:= configRepo.Worktree()

	if err != nil {
		log.Err.Println(err)
		return err
	}

	log.Debug.Println("Git repo updating...")
	err = wt.Pull(configRepoPullOptions)

	if err != nil {
		switch err.Error() {
		case "already up-to-date":
			log.Debug.Println("Git repo already up-to-date.")
		default:
			log.Err.Println(err)
		}
	}

	lastUpdated = time.Now()

	return err
}
