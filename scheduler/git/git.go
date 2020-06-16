package git

import (
	//"fmt"
	"path/filepath"
	"github.com/go-git/go-git/v5"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	gitobject "github.com/go-git/go-git/v5/plumbing/object"

	"github.com/redhat-gpe/scheduler/log"
	"strings"
	"io/ioutil"
	"os"
	"time"
)

var configRepo *git.Repository
var configRepoDir string
var configRepoCloneOptions *git.CloneOptions
var configRepoPullOptions *git.PullOptions

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
			Depth: 1,
			Auth: auth,
		}
		configRepoPullOptions = &git.PullOptions{
			Depth: 1,
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
			Depth: 1,
		}
		configRepoPullOptions = &git.PullOptions{
			Depth: 1,
		}
		configRepo, err = git.PlainClone(dir, false, configRepoCloneOptions)
		if err != nil {
			log.Err.Fatal(err)
		}
	}
	configRepoDir = dir
}

var (
	pullQueue = make(chan bool)
)

// This function watches the channel 'pullQueue' and executes RefreshRepository when there
// is a request with a delay of 10 seconds between each call.
// The goal is to avoid spamming github (or whatever the provider).
func ConsumePullQueue() {
	for {
		select {
		case <- pullQueue:
			RefreshRepository()
			// Empty the queue now that it is refreshed
			pullQueue = make(chan bool)
			time.Sleep(10 * time.Second)
		}
	}
}

func RequestPull() {
	pullQueue <- true
}

// This function refreshes the git Worktree containing the configuration of the scheduler.
// It's basically running the equivalent of 'git pull'.
func RefreshRepository() error {
	// Use WaitGroup to avoid spamming github with 'git pull'
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

	return err
}
