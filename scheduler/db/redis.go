package db


import (
	"github.com/gomodule/redigo/redis"
	"github.com/redhat-gpe/agnostics/log"
	"time"
	"math"
)

var redisURL string

func InitContext(url string) {
	redisURL = url
}

// Dial returns the redis.Conn.
// The context must be set before with InitContext
func Dial() (redis.Conn, error) {
	conn, err := redis.DialURL(redisURL)
	if err != nil {
		log.Err.Println(err)
	}

	return conn, err
}

// DialPubSub is the same as Dial but returns redis.PubSubConn
// The context must be set before with InitContext
func DialPubSub() (redis.PubSubConn, error) {
	conn, err := Dial()

	return redis.PubSubConn{Conn:conn}, err
}

// Reconnect calls Dial until the connection to redis is established.
// This function is blocking and may never end.
func Reconnect() redis.Conn {
	conn, err :=  Dial()
	var wait float64 = 1
	for ; err != nil ; conn, err = Dial() {
		delay := (time.Duration)(math.Pow(2, wait)) * time.Second
		log.Err.Println("Cannot connect to redis. Retrying in", delay, "seconds...")
		time.Sleep(delay)
		if wait < 6 {
			wait = wait + 1
		}
	}

	return conn
}

// ReconnectPubSub is the same as Reconnect except it returns a redis.PubSubConn connection.
func ReconnectPubSub() redis.PubSubConn {
	conn := Reconnect()

	return redis.PubSubConn{Conn: conn}
}
