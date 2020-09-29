package api

import (
	"strconv"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/redhat-gpe/agnostics/internal/config"
	"github.com/redhat-gpe/agnostics/internal/log"
	"github.com/redhat-gpe/agnostics/internal/watcher"
	"github.com/redhat-gpe/agnostics/internal/db"
	"github.com/redhat-gpe/agnostics/internal/api/v1"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"fmt"
)

func v1PostTaintByCloudName(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	functionName := "v1PostTaintByCloudName:"
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	cloudName := params.ByName("cloudname")
	if cloudName == "" {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: 400,
			Message: "Cloud name must be specified in the request",
		})
		return
	}

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
	log.Out.Println(functionName, "Body received: ", string(body))

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
	t := v1.NewTaint()
	if err := dec.Decode(&t); err != io.EOF  && err != nil {
		log.Out.Println(functionName, err)
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: "Error reading data from body. "+err.Error(),
		})
		return
	}

	if t.Key == "" || t.Effect == "" {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: "Taint must have 'key' and 'effect'. ",
		})
		return
	}

	if t.Effect != v1.TaintEffectNoSchedule && t.Effect != v1.TaintEffectPreferNoSchedule {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: "Taint.effect must be 'NoSchedule' or 'PreferNoSchedule'. ",
		})
		return
	}

	clouds := config.GetClouds()
	cloud, ok := clouds[cloudName]
	if ! ok {
		w.WriteHeader(http.StatusNotFound)
		enc.Encode(v1.Error{
			Code: http.StatusNotFound,
			Message: "Cloud Not Found.",
		})
		return
	}
	cloud.Taint(t)
	db.SaveTaints(cloud)
	clouds[cloudName] = cloud
	watcher.RequestTaintSync()
	if err := enc.Encode(cloud); err != nil {
		log.Err.Println(functionName, err)
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(v1.Error{
			Code: http.StatusInternalServerError,
			Message: "Marshal JSON error",
		})
		return
	}

}

func v1DeleteTaintByIndex(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	functionName := "v1DeleteTaint:"
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	cloudName := params.ByName("cloudname")
	if cloudName == "" {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: "Cloud name must be specified in the request",
		})
		return
	}

	clouds := config.GetClouds()
	cloud, ok := clouds[cloudName]
	if ! ok {
		w.WriteHeader(http.StatusNotFound)
		enc.Encode(v1.Error{
			Code: http.StatusNotFound,
			Message: "Cloud Not Found.",
		})
		return
	}

	taintIndex, err := strconv.Atoi(params.ByName("taintindex"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(v1.Error{
			Code: http.StatusInternalServerError,
			Message: "Can't convert index to integer",
		})
		return
	}

	if len(cloud.Taints) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: fmt.Sprintf("Cloud %s has no taint.", cloud.Name),
		})
		return
	}


	if taintIndex >= len(cloud.Taints) || taintIndex < 0{
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: fmt.Sprintf("Taint index out of range (must be 0..%d)", len(cloud.Taints)-1),
		})
		return
	}

	cloud.Taints = append(cloud.Taints[:taintIndex], cloud.Taints[taintIndex+1:]...)
	clouds[cloudName] = cloud
	db.SaveTaints(cloud)
	watcher.RequestTaintSync()
	if err := enc.Encode(cloud); err != nil {
		log.Err.Println(functionName, err)
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(v1.Error{
			Code: http.StatusInternalServerError,
			Message: "Marshal JSON error",
		})
		return
	}
}

func v1DeleteTaintByCloudName(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	functionName := "v1DeleteTaintByCloudName:"
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	cloudName := params.ByName("cloudname")
	if cloudName == "" {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: "Cloud name must be specified in the request",
		})
		return
	}
	clouds := config.GetClouds()
	cloud, ok := clouds[cloudName]
	if ! ok {
		w.WriteHeader(http.StatusNotFound)
		enc.Encode(v1.Error{
			Code: http.StatusNotFound,
			Message: "Cloud Not Found.",
		})
		return
	}

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
	log.Out.Println(functionName, "Body received: ", string(body))

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
	t := v1.NewTaint()
	if err := dec.Decode(&t); err != io.EOF  && err != nil {
		log.Out.Println(functionName, err)
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: "Error reading data from body. "+err.Error(),
		})
		return
	}

	if t.Key == "" || t.Effect == "" {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: "Taint must have 'key' and 'effect'. ",
		})
		return
	}

	if t.Effect != v1.TaintEffectNoSchedule && t.Effect != v1.TaintEffectPreferNoSchedule {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: "Taint.effect must be 'NoSchedule' or 'PreferNoSchedule'. ",
		})
		return
	}

	if len(cloud.Taints) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: fmt.Sprintf("Cloud %s has no taint.", cloud.Name),
		})
		return
	}

	// Remove the taint, build new taint list
	result := []v1.Taint{}
	for _, taint := range cloud.Taints {
		if ! taint.MatchTaint(t) {
			result = append(result, taint)
		}
	}
	cloud.Taints = result
	clouds[cloudName] = cloud
	db.SaveTaints(cloud)
	watcher.RequestTaintSync()
	if err := enc.Encode(cloud); err != nil {
		log.Err.Println(functionName, err)
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(v1.Error{
			Code: http.StatusInternalServerError,
			Message: "Marshal JSON error",
		})
		return
	}
}

func v1DeleteTaintsByCloudName(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	functionName := "v1DeleteTaintByCloudName:"
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	cloudName := params.ByName("cloudname")
	if cloudName == "" {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: "Cloud name must be specified in the request",
		})
		return
	}

	clouds := config.GetClouds()
	cloud, ok := clouds[cloudName]
	if ! ok {
		w.WriteHeader(http.StatusNotFound)
		enc.Encode(v1.Error{
			Code: http.StatusNotFound,
			Message: "Cloud Not Found.",
		})
		return
	}

	if len(cloud.Taints) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(v1.Error{
			Code: http.StatusBadRequest,
			Message: fmt.Sprintf("Cloud %s has no taint.", cloud.Name),
		})
		return
	}

	cloud.Taints = []v1.Taint{}
	clouds[cloudName] = cloud
	db.SaveTaints(cloud)
	watcher.RequestTaintSync()
	if err := enc.Encode(cloud); err != nil {
		log.Err.Println(functionName, err)
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(v1.Error{
			Code: http.StatusInternalServerError,
			Message: "Marshal JSON error",
		})
		return
	}
}
