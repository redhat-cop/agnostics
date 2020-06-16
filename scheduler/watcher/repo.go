package watcher

import(
	"github.com/redhat-gpe/scheduler/git"
	"github.com/redhat-gpe/scheduler/config"
	"time"
)

var (
	pullQueue = make(chan bool)
)


func RequestPull() {
	pullQueue <- true
}

// This function watches the channel 'pullQueue' and executes RefreshRepository when there
// is a request with a delay of 10 seconds between each call.
// The goal is to avoid spamming github (or whatever the provider).
func ConsumePullQueue() {
	for {
		select {
		case <- pullQueue:
			git.RefreshRepository()
			config.Load()
			// Empty the queue now that it is refreshed
			pullQueue = make(chan bool)
			time.Sleep(10 * time.Second)
		}
	}
}
