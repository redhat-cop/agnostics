package watcher

import(
	"github.com/redhat-gpe/scheduler/git"
	"github.com/redhat-gpe/scheduler/log"
	"github.com/redhat-gpe/scheduler/config"
	"github.com/redhat-gpe/scheduler/db"
	"github.com/gomodule/redigo/redis"
)

func RequestPull() {
	conn :=  db.Dial()
	defer conn.Close()

	conn.Do("PUBLISH", "repoMQ", "pull")
}

// This function watches the message Queue 'repoMQ' in redis
// and executes RefreshRepository when there is a request with
// a delay of 10 seconds between each call.
// The goal is to avoid spamming github (or whatever the provider).
func ConsumePullQueue() {
	conn :=  redis.PubSubConn{Conn:db.Dial()}
	defer conn.Close()

	conn.Subscribe("repoMQ")
	for {
		switch v := conn.Receive().(type) {
		case redis.Message:
			log.Debug.Printf("%s: message: %s\n", v.Channel, v.Data)
			if err := git.RefreshRepository(); err == nil {
				config.Load()
			}
		case redis.Subscription:
			log.Debug.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
			continue
		case error:
			log.Debug.Println(v)
		}
	}
}
