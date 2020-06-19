package db


import (
	"github.com/gomodule/redigo/redis"
	"github.com/redhat-gpe/scheduler/log"
)

var redisURL string

func InitContext(url string) {
	redisURL = url
}

func Dial() redis.Conn {
	conn, err := redis.DialURL(redisURL)
	if err != nil {
		log.Err.Println(err)
	}

	return conn
}
