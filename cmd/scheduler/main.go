package main

import(
	"flag"
	"github.com/redhat-gpe/agnostics/internal/api"
	"github.com/redhat-gpe/agnostics/internal/console"
	"github.com/redhat-gpe/agnostics/internal/config"
	"github.com/redhat-gpe/agnostics/internal/git"
	"github.com/redhat-gpe/agnostics/internal/log"
	"github.com/redhat-gpe/agnostics/internal/watcher"
	"github.com/redhat-gpe/agnostics/internal/db"
	"os"
)

// Flags
var debugFlag bool
var repositoryURL string
var sshPrivateKey string
var redisURL string
var templateDir string
var apiAddress string
var consoleAddress string
var apiAuth bool
var apiHtpasswd string

func parseFlags() {
	flag.StringVar(&repositoryURL, "git-url", "git@github.com:redhat-gpe/scheduler-config.git", "The URL of the git repository where the scheduler will find its configuration. SSH is assumed, unless the URL starts with 'http'.\nEnvironment variable: GIT_URL\n")
	flag.StringVar(&sshPrivateKey, "git-ssh-private-key", "", "The path of the SSH private key used to authenticate to the git repository. Used only when 'git-url' is an SSH URL.\nEnvironment variable: GIT_SSH_PRIVATE_KEY\n")
	flag.StringVar(&redisURL, "redis-url", "redis://localhost:6379", "The URL to access redis. The format is described by the IANA specification for the scheme, see https://www.iana.org/assignments/uri-schemes/prov/redis\nEnvironment variable: REDIS_URL\n")
	flag.BoolVar(&debugFlag, "debug", false, "Debug mode.\nEnvironment variable: DEBUG\n")
	flag.StringVar(&templateDir, "template-dir", "templates", "The directory containing the golang templates for the Console.\nEnvironment variable: TEMPLATE_DIR\n")
	flag.StringVar(&apiAddress, "api-addr", ":8080", "The address API listens to.\nEnvironment variable: API_ADDR\n")
	flag.StringVar(&consoleAddress, "console-addr", ":8081", "The address the Console listens to.\nEnvironment variable: CONSOLE_ADDR\n")
	flag.BoolVar(&apiAuth, "api-auth", true, "Enable authentication for the API.\nEnvironment variable: API_AUTH  ('true' or 'false')\n")
	flag.StringVar(&apiHtpasswd, "api-htpasswd", "api-htpasswd", "The path of the htpasswd file to use for authentication for the API.\nEnvironment variable: API_HTPASSWD\n")

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
	if e := os.Getenv("API_ADDR"); e != "" {
		apiAddress = e
	}
	if e := os.Getenv("CONSOLE_ADDR"); e != "" {
		consoleAddress = e
	}
	if e := os.Getenv("TEMPLATE_DIR"); e != "" {
		templateDir = e
	}
	if e := os.Getenv("API_AUTH"); e == "false" {
		apiAuth = false
	}
	if e := os.Getenv("API_HTPASSWD"); e != "" {
		apiHtpasswd = e
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
	go watcher.ConsumeTaintSyncQueue()
	config.Load()
	go console.Serve(templateDir, consoleAddress)
	api.Serve(apiAddress, apiAuth, apiHtpasswd)
}
