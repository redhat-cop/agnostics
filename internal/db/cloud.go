package db

import(
	"encoding/json"
	"fmt"
	"errors"
	"github.com/redhat-gpe/agnostics/internal/log"
	"github.com/redhat-gpe/agnostics/internal/api/v1"
	"github.com/gomodule/redigo/redis"
)

func SaveTaints(cloud v1.Cloud) error {
	conn, err := Dial()
	if err != nil {
		log.Err.Println("Cannot connect to redis:", err)
		return err
	}
	defer conn.Close()

	jsonText, err :=  json.Marshal(cloud.Taints)
	if reply, err := conn.Do("JSON.SET", "taints:"+cloud.Name, ".", jsonText); err != nil {
		log.Err.Println("cloud.SaveTaints(taints:", cloud.Name,")", err)
		return err
	} else {
		log.Debug.Println("cloud.SaveTaints(taints:",cloud.Name, ")", reply)
		return nil
	}

}

func ReloadAllTaints(clouds map[string]v1.Cloud) error {
	functionName := "ReloadAllTaints:"
	for k, v := range clouds {
		taints, err := getTaints(v)
		if err != nil {
			log.Err.Println(functionName, err)
		}
		v.Taints = taints
		clouds[k] = v
	}
	log.Out.Println(functionName, "OK")
	return nil
}

// Error when the placement is not found using Uuid
var ErrTaintsNotFound = errors.New("taints not found")

func getTaints(cloud v1.Cloud) ([]v1.Taint, error) {
	conn, err := Dial()
	if err != nil {
		log.Err.Println("Cannot connect to redis:", err)
		return []v1.Taint{}, err
	}
	defer conn.Close()

	key := fmt.Sprintf("taints:%s", cloud.Name)

	if reply, err := redis.Bytes(conn.Do("JSON.GET", key)); err != nil {
		if err == redis.ErrNil {
			return []v1.Taint{}, nil
		}
		return []v1.Taint{}, err
	} else if reply == nil {
		return []v1.Taint{}, nil
	} else {
		var taints []v1.Taint
		if err := json.Unmarshal(reply, &taints); err != nil {
			return []v1.Taint{}, err
		}
		return taints, nil
	}
}
