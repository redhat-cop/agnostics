package main

import(
	"flag"
	"github.com/redhat-gpe/scheduler/api"
	"github.com/redhat-gpe/scheduler/config"
	"github.com/redhat-gpe/scheduler/git"
	"github.com/redhat-gpe/scheduler/log"
	"github.com/redhat-gpe/scheduler/watcher"
	"github.com/redhat-gpe/scheduler/db"
	"os"
)

// Flags
var debugFlag bool
var repositoryURL string
var sshPrivateKey string
var redisURL string

func parseFlags() {
	flag.StringVar(&repositoryURL, "git-url", "git@github.com:redhat-gpe/scheduler-config.git", "The URL of the git repository where the scheduler will find its configuration. SSH is assumed, unless the URL starts with 'http'. To set git-url you can also use the environment variable REPOSITORY_URL. If both are set, environment variable prevails.")
	flag.StringVar(&sshPrivateKey, "ssh-private-key", "", "The SSH private key used to authenticate to the git repository. Used only when 'git-url' is an SSH URL. Environment variable: SSH_PRIVATE_KEY.")
	flag.StringVar(&redisURL, "redis-url", "redis://localhost:6379", "The URL to access redis. Environment variable: REDIS_URL. The format is described by the IANA specification for the scheme, see https://www.iana.org/assignments/uri-schemes/prov/redis")
	flag.BoolVar(&debugFlag, "debug", false, "Debug mode. Environment variable: DEBUG.")
	flag.Parse()
	if e := os.Getenv("REPOSITORY_URL"); e != "" {
		repositoryURL = e
	}
	if e := os.Getenv("SSH_PRIVATE_KEY"); e != "" {
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
