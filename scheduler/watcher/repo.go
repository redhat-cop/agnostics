package watcher

import(
	"github.com/redhat-gpe/scheduler/git"
	"github.com/redhat-gpe/scheduler/log"
	"github.com/redhat-gpe/scheduler/config"
	"github.com/redhat-gpe/scheduler/db"
	"github.com/gomodule/redigo/redis"
)

func RequestPull() {
	conn, err :=  db.Dial()
	if err != nil {
		log.Err.Println("Cannot connect to redis. Repo not updated.")
		return
	}
	defer conn.Close()

	conn.Do("PUBLISH", "repoMQ", "pull")
}

// ConsumePullQueue function watches the message Queue 'repoMQ' in redis
// and executes RefreshRepository when there is a request with
// a delay of 10 seconds between each call.
// The goal is to avoid spamming github (or whatever the provider).
func ConsumePullQueue() {
	conn :=  db.ReconnectPubSub()
	defer conn.Close()

	conn.Subscribe("repoMQ")
	defer conn.Unsubscribe("repoMQ")
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
			conn = db.ReconnectPubSub()
			conn.Subscribe("repoMQ")
		}
	}
}
