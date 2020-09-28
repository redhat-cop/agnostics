package api

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/redhat-gpe/agnostics/internal/config"
	"github.com/redhat-gpe/agnostics/internal/git"
	"github.com/redhat-gpe/agnostics/internal/log"
	"github.com/redhat-gpe/agnostics/internal/modules"
	"github.com/redhat-gpe/agnostics/internal/watcher"
	"github.com/redhat-gpe/agnostics/internal/api/v1"
	"github.com/redhat-gpe/agnostics/internal/placement"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func v1GetClouds(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	clouds := []v1.Cloud{}
	for _, v := range config.GetClouds() {
		clouds = append(clouds, v)
	}
	if err := enc.Encode(clouds); err != nil {
		log.Err.Println("GET clouds", err)
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(v1.Error{
			Code: 500,
			Message: "Error reading clouds from config.",
		})
	}
}

func v1PostSchedule(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Err.Println("POST schedule", err)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: "Error reading body from request.",
		})
		return
	}
	log.Out.Println("POST schedule, Body received: ", string(body))

	if ! json.Valid([]byte(body)) {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: "Body is not valid JSON.",
		})
		return
	}

	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.DisallowUnknownFields()
	scheduleQuery := new(v1.ScheduleQuery)
	if err := dec.Decode(scheduleQuery); err != io.EOF  && err != nil {
		log.Out.Println("POST schedule", err)
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: "Error reading data from body. "+err.Error(),
		})
		return
	}

	if scheduleQuery.UUID == "" {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: "uuid must be provided",
		})
		return
	}

	if len(scheduleQuery.Annotations) > 0 {
		for k, v := range scheduleQuery.Annotations {
			if k == "" || v == "" {
				w.WriteHeader(http.StatusBadRequest)
				enc.Encode(v1.Error{
					Code: http.StatusBadRequest,
					Message: "Annotations keys and values cannot be empty string.",
				})
				return
			}
		}
	}

	if _, err := placement.Get(scheduleQuery.UUID) ; err != placement.ErrPlacementNotFound {
		if err == nil {
			w.WriteHeader(http.StatusBadRequest)
			enc.Encode(v1.Error{
				Code: http.StatusBadRequest,
				Message: "This service uuid already has a placement",
			})
			return

		}
		// Else something went wrong
		log.Err.Println("POST schedule", err)
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(v1.Error{
			Code: http.StatusInternalServerError,
			Message: "Internal Server Error",
		})
		return
	}

	policy := config.GetPolicy()
	clouds := []v1.Cloud{}
	for _, c := range config.GetClouds() {
		clouds = append(clouds, c)
	}

	for _, predicate := range policy.Predicates {
		switch predicate.Name {
		case "LabelPredicates":
			clouds = modules.LabelPredicates(clouds, scheduleQuery.CloudSelector)
		case "TaintPredicates":
			clouds = modules.TaintPredicates(clouds, scheduleQuery.Tolerations)
		default:
			log.Err.Println("Predicate", predicate.Name, "not supported")
		}
	}

	for _, priority := range policy.Priorities {
		switch priority.Name {
		case "LabelPriorities":
			clouds = modules.LabelPriorities(clouds, scheduleQuery.CloudPreference, priority.Weight)
		case "TaintPriorities":
			clouds = modules.TaintPriorities(clouds, scheduleQuery.Tolerations, priority.Weight)
		default:
			log.Err.Println("Priority", priority.Name, "not supported")
		}
	}

	if len(clouds) == 0 {
		log.Out.Println("POST schedule", err)
		w.WriteHeader(http.StatusNotFound)
		enc.Encode(v1.Error{
			Code: http.StatusNotFound,
			Message: "No cloud found.",
		})
		return
	}
	// pick the first one
	result := v1.Placement{
		UUID: scheduleQuery.UUID,
		Cloud: clouds[0],
		CreationTimestamp: time.Now().UTC().Round(time.Second),
		Annotations: scheduleQuery.Annotations,
	}
	placement.Save(result)
	if err := enc.Encode(result) ; err != nil {
		log.Err.Println("POST schedule", err)
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(v1.Error{
			Code: http.StatusInternalServerError,
			Message: "Error encoding data to JSON",
		})
	}
}

func v1GetPlacements(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	log.Debug.Println("GET all placements")
	if p, err := placement.GetAll(0) ; err == nil {
		if err := enc.Encode(p); err != nil {
			log.Err.Println("GET placement", err)
			w.WriteHeader(http.StatusInternalServerError)
			enc.Encode(v1.Error{
				Code: 500,
				Message: "Marshal JSON error",
			})
			return
		}
	} else {
		log.Err.Println("GET placement", err)
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(v1.Error{
			Code: 500,
			Message: "Internal Server Error",
		})
	}
}

func v1GetPlacement(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	uuid := params.ByName("uuid")
	if uuid == "" {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: 400,
			Message: "UUID must be specified in the request",
		})
		return
	}
	if p, err := placement.Get(uuid) ; err == nil {
		log.Debug.Println("GET placement", p)
		if err := enc.Encode(p); err != nil {
			log.Err.Println("GET placement", err)
			w.WriteHeader(http.StatusInternalServerError)
			enc.Encode(v1.Error{
				Code: 500,
				Message: "Marshal JSON error",
			})
			return
		}
	} else if err == placement.ErrPlacementNotFound {
		w.WriteHeader(http.StatusNotFound)
		enc.Encode(v1.Error{
			Code: 404,
			Message: "Placement not found.",
		})
	} else {
		log.Err.Println("GET placement", err)
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(v1.Error{
			Code: 500,
			Message: "Internal Server Error",
		})
	}
}

func v1DeletePlacement(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	uuid := params.ByName("uuid")
	if uuid == "" {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: 400,
			Message: "UUID must be specified in the request",
		})
		return
	}
	if _, err := placement.Get(uuid) ; err == nil {
		log.Debug.Println("DELETE placement", uuid)
		if err := placement.Delete(uuid) ; err == nil {
			enc.Encode(v1.Message{
				Message: "placement deleted",
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Err.Println("DELETE placement", err)
			enc.Encode(v1.Error{
				Code: 500,
				Message: "error deleting placement",
			})
			return
		}
	} else if err == placement.ErrPlacementNotFound {
		w.WriteHeader(http.StatusNotFound)
		enc.Encode(v1.Error{
			Code: 404,
			Message: "Placement not found.",
		})
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		log.Err.Println("DELETE placement", err)
		enc.Encode(v1.Error{
			Code: 500,
			Message: "Internal Server Error",
		})
	}
}

func v1GetCloudByName(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	clouds := config.GetClouds()
	if val, ok := clouds[params.ByName("name")] ; ok {
		if err := enc.Encode(val); err != nil {
			log.Err.Println("GET cloud", err)
			w.WriteHeader(http.StatusInternalServerError)
			enc.Encode(v1.Error{
				Code: 500,
				Message: "Marshal JSON error",
			})
		}
		return

	}
	w.WriteHeader(http.StatusNotFound)
	enc.Encode(v1.Error{
		Code: 404,
		Message: "Cloud not found.",
	})
}

func v1GetRepository(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	if gitCommit, err := v1.NewGitCommit(git.GetRepo()); err == nil {
		enc.Encode(gitCommit)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(v1.Error{
			Code: http.StatusInternalServerError,
			Message: "ERROR while retrieving Git HEAD information.",
		})
	}
}

func v1PullRepository(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	go watcher.RequestPull()
	enc.Encode(v1.Message{
		Message: "Request to update git repository received.",
	})
}

func v1PutCounters(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	log.Out.Println("Refresh all counters")
	err := placement.RefreshAllCounters()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Err.Println(err)
		enc.Encode(v1.Error{
			Code: http.StatusInternalServerError,
			Message: "ERROR while refreshing counters.",
		})

	}
	enc.Encode(v1.Message{
		Message: "All counters updated",
	})
}
