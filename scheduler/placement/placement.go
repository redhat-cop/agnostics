package placement

import(
	"github.com/redhat-gpe/scheduler/db"
	"github.com/redhat-gpe/scheduler/log"
	"github.com/redhat-gpe/scheduler/api/v1"
	"github.com/gomodule/redigo/redis"
	"errors"
	"encoding/json"
)

// Error when the placement is not found using Uuid
var ErrPlacementNotFound = errors.New("placement not found")

func get(key string) (v1.Placement, error) {
	conn, err := db.Dial()
	defer conn.Close()
	if err != nil {
		log.Err.Println("Cannot connect to redis:", err)
		return v1.Placement{}, err
	}

	if reply, err := redis.Bytes(conn.Do("JSON.GET", key)); err != nil {
		if err == redis.ErrNil {
			return v1.Placement{}, ErrPlacementNotFound
		}
		return v1.Placement{}, err
	} else if reply == nil {
		return v1.Placement{}, ErrPlacementNotFound
	} else {
		log.Debug.Println("reply Get(", key ,")=", string(reply))

		var p v1.Placement
		if err := json.Unmarshal(reply, &p); err != nil {
			return v1.Placement{}, err
		}
		return p, nil
	}
}

// Get retrives a placement from the DB.
func Get(uuid string) (v1.Placement, error) {
	return get("placement:"+uuid)
}

// Get retrives a placement from the DB.
func GetAll() ([]v1.Placement, error) {
	conn, err := db.Dial()
	result := []v1.Placement{}
	defer conn.Close()
	if err != nil {
		log.Err.Println("Cannot connect to redis:", err)
		return []v1.Placement{}, err
	}

	keys, err := redis.Strings(conn.Do("KEYS", "placement:*"))
	if err != nil {
		log.Err.Println("placement.Get error", err)
		return []v1.Placement{}, err
	}
	for _, key := range keys {
		log.Debug.Println(key)
		if p, err := get(key); err == nil {
			result = append(result, p)
		}
	}
	return result, err
}

// Save saves a placement in the database.
func Save(p v1.Placement) error {
	conn, err := db.Dial()
	defer conn.Close()
	if err != nil {
		log.Err.Println("Cannot connect to redis:", err)
		return err
	}

	jsonText, err :=  json.Marshal(p)

	if reply, err := conn.Do("JSON.SET", "placement:"+p.UUID, ".", jsonText); err != nil {
		log.Err.Println("placement.Set(", p.UUID,")", err)
		return err
	} else if reply == nil {
		log.Err.Println("placement.Set(", p.UUID, ") reply is nil")
		return nil
	} else {
		log.Debug.Println("placement.Set(", p.UUID,")", reply)
		return nil
	}
}

// Delete deletes a placement from the database.

func Delete(uuid string) error {
	conn, err := db.Dial()
	defer conn.Close()
	if err != nil {
		log.Err.Println("Cannot connect to redis:", err)
		return err
	}

	if reply, err := conn.Do("JSON.DEL", "placement:"+uuid); err != nil {
		log.Debug.Println("placement.Delete(", uuid ,")=", reply)
		if err == redis.ErrNil {
			return ErrPlacementNotFound
		}
		return err
	} else if reply == nil {
		return ErrPlacementNotFound
	} else {
		log.Debug.Println("reply Delete(", uuid ,")=", reply)
		return nil
	}
}
