package main

import(
	"flag"
	"github.com/redhat-gpe/agnostics/api"
	"github.com/redhat-gpe/agnostics/config"
	"github.com/redhat-gpe/agnostics/git"
	"github.com/redhat-gpe/agnostics/log"
	"github.com/redhat-gpe/agnostics/watcher"
	"github.com/redhat-gpe/agnostics/db"
	"os"
)

// Flags
var debugFlag bool
var repositoryURL string
var sshPrivateKey string
var redisURL string

func parseFlags() {
	flag.StringVar(&repositoryURL, "git-url", "git@github.com:redhat-gpe/scheduler-config.git", "The URL of the git repository where the scheduler will find its configuration. SSH is assumed, unless the URL starts with 'http'.\nEnvironment variable: GIT_URL\n")
	flag.StringVar(&sshPrivateKey, "git-ssh-private-key", "", "The path of the SSH private key used to authenticate to the git repository. Used only when 'git-url' is an SSH URL.\nEnvironment variable: GIT_SSH_PRIVATE_KEY\n")
	flag.StringVar(&redisURL, "redis-url", "redis://localhost:6379", "The URL to access redis. The format is described by the IANA specification for the scheme, see https://www.iana.org/assignments/uri-schemes/prov/redis\nEnvironment variable: REDIS_URL\n")
	flag.BoolVar(&debugFlag, "debug", false, "Debug mode.\nEnvironment variable: DEBUG\n")
	flag.Parse()
	if e := os.Getenv("GIT_URL"); e != "" {
		repositoryURL = e
	}
	if e := os.Getenv("GIT_SSH_PRIVATE_KEY"); e != "" {
		sshPrivateKey = e
	}
	if e := os.Getenv("REDIS_URL"); e != "" {
		redisURL = e
	}
	if e := os.Getenv("DEBUG"); e != "" && e != "false" {
		debugFlag = true
	}
}

func main() {
	parseFlags()
	log.InitLoggers(debugFlag)
	db.InitContext(redisURL)
	git.CloneRepository(repositoryURL, sshPrivateKey)
	go watcher.ConsumePullQueue()
	config.Load()
	api.Serve()
}
