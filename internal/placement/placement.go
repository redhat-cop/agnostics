package placement

import(
	"github.com/redhat-gpe/agnostics/internal/db"
	"github.com/redhat-gpe/agnostics/internal/log"
	"github.com/redhat-gpe/agnostics/internal/api/v1"
	"github.com/gomodule/redigo/redis"
	"errors"
	"encoding/json"
)

// Error when the placement is not found using Uuid
var ErrPlacementNotFound = errors.New("placement not found")

func get(key string) (v1.Placement, error) {
	conn, err := db.Dial()
	if err != nil {
		log.Err.Println("Cannot connect to redis:", err)
		return v1.Placement{}, err
	}
	defer conn.Close()

	if reply, err := redis.Bytes(conn.Do("JSON.GET", key)); err != nil {
		if err == redis.ErrNil {
			return v1.Placement{}, ErrPlacementNotFound
		}
		return v1.Placement{}, err
	} else if reply == nil {
		return v1.Placement{}, ErrPlacementNotFound
	} else {
		var p v1.Placement
		if err := json.Unmarshal(reply, &p); err != nil {
			return v1.Placement{}, err
		}
		if p.CreationTimestamp.IsZero() && ! p.Date.IsZero() {
			// Probably an old record that doesn't have creation_timestamp field set.
			// Use old deprecated field 'date'
			p.CreationTimestamp = p.Date
		}
		return p, nil
	}
}

// Get retrives a placement from the DB.
func Get(uuid string) (v1.Placement, error) {
	return get("placement:"+uuid)
}

// Get retrives a placement from the DB.
// The 'count' parameter is the maximum number of placements to be returned.
// Set 'count' to  0 if you want the function to return all placements without limit.
func GetAll(count int) ([]v1.Placement, error) {
	result := []v1.Placement{}
	conn, err := db.Dial()
	if err != nil {
		log.Err.Println("Cannot connect to redis:", err)
		return []v1.Placement{}, err
	}
	defer conn.Close()

	// here we'll store our iterator value
	iter := 0

	// this will store the keys of each iteration
	var keys []string
	for {

		// we scan with our iter offset, starting at 0
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", "placement:*"))
		if err != nil {
			log.Err.Println("placement.Get error", err)
			return []v1.Placement{}, err
		}
		// now we get the iter and the keys from the multi-bulk reply
		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)

		keys = append(keys, k...)

		// check if we need to stop...
		if count != 0 && len(keys) > count {
			break
		}
		if iter == 0 {
			break
		}
	}
	for _, key := range keys {
		if p, err := get(key); err == nil {
			result = append(result, p)
		}
	}
	return result, err
}

// Save saves a placement in the database.
func Save(p v1.Placement) error {
	conn, err := db.Dial()
	if err != nil {
		log.Err.Println("Cannot connect to redis:", err)
		return err
	}
	defer conn.Close()

	newEntry := false
	pOld, err := Get(p.UUID)
	if err == ErrPlacementNotFound {
		newEntry = true
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
		// Update counter
		countPlacementsByCloud(conn, "INCR", p.Cloud.Name)
		countPlacementsByCloud(conn, "INCR", "all")
		if newEntry == false {
			countPlacementsByCloud(conn, "DECR", pOld.Cloud.Name)
			countPlacementsByCloud(conn, "DECR", "all")
		}

		return nil
	}
}

func countPlacementsByCloud(conn redis.Conn, command string, name string) error {
	if reply, err := conn.Do(command, "counter:placements:"+name); err != nil {
		log.Err.Println(command, "(counter:placements:", name,")", err)
		return err
	} else {
		log.Debug.Println(command, "(counter:placements:", name,")", reply)
		return nil
	}
}

// GetCountPlacementsByCloud return the counter for that cloud name.
func GetCountPlacementsByCloud(name string) (string, error) {
	conn, err := db.Dial()
	if err != nil {
		log.Err.Println("Cannot connect to redis:", err)
		return "", err
	}
	defer conn.Close()

	if reply, err := redis.String(conn.Do("GET", "counter:placements:"+name)); err != nil {
		if err.Error() == "redigo: nil returned" {
			return "0", nil
		}
		log.Err.Println("GET", "(counter:placements:", name,")", err)
		return reply,  err
	} else {
		return reply, nil
	}
}

// RefreshAllCounters calculates and refreshes all the counters
func RefreshAllCounters() error {
	placements, err := GetAll(0)
	if err != nil {
		return err
	}

	byCloud := map[string]int{}

	for _, p := range placements {
		byCloud[p.Cloud.Name] = byCloud[p.Cloud.Name] + 1
	}

	conn, err := db.Dial()
	if err != nil {
		log.Err.Println("Cannot connect to redis:", err)
		return err
	}
	defer conn.Close()

	for k, v := range byCloud {
		if _, err := conn.Do("SET", "counter:placements:"+k, v); err != nil {
			log.Err.Println("SET", "(counter:placements:", k,")", err)
			return err
		}
	}
	if _, err := conn.Do("SET", "counter:placements:all", len(placements)); err != nil {
		log.Err.Println("SET", "(counter:placements:all)", err)
		return err
	}
	return nil
}

// Delete deletes a placement from the database.

func Delete(uuid string) error {
	conn, err := db.Dial()
	if err != nil {
		log.Err.Println("Cannot connect to redis:", err)
		return err
	}
	defer conn.Close()

	p, err := Get(uuid)
	if err == ErrPlacementNotFound {
		return ErrPlacementNotFound
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
		countPlacementsByCloud(conn, "DECR", p.Cloud.Name)
		countPlacementsByCloud(conn, "DECR", "all")
		return nil
	}
}
