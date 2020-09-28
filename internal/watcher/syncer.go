package watcher

import(
	"github.com/redhat-gpe/agnostics/internal/log"
	"github.com/redhat-gpe/agnostics/internal/config"
	"github.com/redhat-gpe/agnostics/internal/db"
	"github.com/gomodule/redigo/redis"
)

func RequestTaintSync() {
	conn, err :=  db.Dial()
	if err != nil {
		log.Err.Println("Cannot connect to redis. Repo not updated.")
		return
	}
	defer conn.Close()

	conn.Do("PUBLISH", "taintMQ", "sync")
}

// ConsumeTaintSyncQueue function watches the message Queue 'taintMQ' in redis
// and executes RefreshTaints when there is a request.
func ConsumeTaintSyncQueue() {
	conn :=  db.ReconnectPubSub()
	defer conn.Close()

	conn.Subscribe("taintMQ")
	defer conn.Unsubscribe("taintMQ")
	for {
		switch v := conn.Receive().(type) {
		case redis.Message:
			log.Debug.Printf("channel %s: message: %s\n", v.Channel, v.Data)
			db.ReloadAllTaints(config.GetClouds())
		case redis.Subscription:
			log.Debug.Printf("channel %s: %s %d\n", v.Channel, v.Kind, v.Count)
			continue
		case error:
			log.Debug.Println(v)
			conn = db.ReconnectPubSub()
			conn.Subscribe("taintMQ")
		}
	}
}
