package main

import(
	"flag"
	"github.com/redhat-gpe/scheduler/api"
	"github.com/redhat-gpe/scheduler/config"
	"github.com/redhat-gpe/scheduler/git"
	"github.com/redhat-gpe/scheduler/log"
)

// Flags
var debugFlag bool
var repositoryURL string
var sshPrivateKey string

func parseFlags() {
	flag.BoolVar(&debugFlag, "debug", false, "Debug mode")
	flag.StringVar(&repositoryURL, "git-url", "git@github.com:redhat-gpe/scheduler-config.git", "The URL of the git repository where the scheduler will find its configuration. SSH is assumed, unless the URL starts with 'http'.")
	flag.StringVar(&sshPrivateKey, "ssh-private-key", "", "The SSH private key used to authenticate to the git repository. Used only when 'git-url' is an SSH URL.")
	flag.Parse()
}

func main() {
	parseFlags()
	log.InitLoggers(debugFlag)
	git.CloneRepository(repositoryURL, sshPrivateKey)
	go git.ConsumePullQueue()
	config.Load()
	api.Serve()
}
