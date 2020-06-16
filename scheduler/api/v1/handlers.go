package v1

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/redhat-gpe/scheduler/config"
	"github.com/redhat-gpe/scheduler/git"
	"github.com/redhat-gpe/scheduler/log"
	"github.com/redhat-gpe/scheduler/modules"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func GetClouds(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	jsonResult, err := json.MarshalIndent(config.GetClouds(), "", "  ")
	if err != nil {
		log.Err.Println(err)
		errorMessage := Error{
			Code: 1,
			Message: "Error reading clouds from config.",
		}
		jsonError, _ := json.MarshalIndent(errorMessage, "", " ")
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, string(jsonError))
	} else {
		io.WriteString(w, string(jsonResult))
	}
}

func Schedule(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Err.Println(err)
		errorMessage := Error{
			Code: http.StatusBadRequest,
			Message: "Error reading body from request.",
		}
		jsonError, _ := json.MarshalIndent(errorMessage, "", " ")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, string(jsonError))
		return
	}
	log.Debug.Println(string(body))

	if ! json.Valid([]byte(body)) {
		errorMessage := Error{
			Code: http.StatusBadRequest,
			Message: "Body is not valid JSON.",
		}
		jsonError, _ := json.MarshalIndent(errorMessage, "", " ")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, string(jsonError))
		return
	}


	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.DisallowUnknownFields()

	t := new(CloudQuery)
	log.Debug.Println(t, "bla", t.CloudSelector)
	if err := dec.Decode(t); err != io.EOF  && err != nil {
		log.Err.Println(err)
		errorMessage := Error{
			Code: http.StatusBadRequest,
			Message: "Error reading data from body. "+err.Error(),
		}
		jsonError, _ := json.MarshalIndent(errorMessage, "", " ")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, string(jsonError))
		return
	}

	log.Debug.Println(t, "bla", t.CloudSelector)

	clouds := modules.LabelPredicates(config.GetClouds(), t.CloudSelector)
	result := modules.LabelPriorities(clouds, t.CloudPreference)

	if len(result) == 0 {
		log.Err.Println(err)
		errorMessage := Error{
			Code: 404,
			Message: "No cloud found.",
		}
		jsonError, _ := json.MarshalIndent(errorMessage, "", " ")
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, string(jsonError))
		return
	}

	// Pick the first one
	jsonResult, err := json.MarshalIndent(result[0], "", "  ")
	if err != nil {
		log.Err.Println(err)
		errorMessage := Error{
			Code: 4,
			Message: "Error marshaling cloud from config to JSON.",
		}
		jsonError, _ := json.MarshalIndent(errorMessage, "", " ")
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, string(jsonError))
	} else {
		io.WriteString(w, string(jsonResult))
	}
}

func GetCloudByName(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	clouds := config.GetClouds()
	if val, ok := clouds[params.ByName("name")] ; ok {
		json, err := json.MarshalIndent(val, "", "  ")
		if err != nil {
			log.Err.Println(err)
			io.WriteString(w, "Error\n")
		} else {
			io.WriteString(w, string(json))
		}

	} else {
		errorMessage := Error{
			Code: 404,
			Message: "Cloud not found.",
		}
		jsonError, _ := json.MarshalIndent(errorMessage, "", " ")
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, string(jsonError))
	}
}

func GetRepository(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if head, err := git.GetRepoHeadCommit() ; err == nil {
		m := GitCommit{
			Hash: head.Hash.String(),
			Author: head.Author.Name,
			Date: head.Author.When.UTC().Format(time.RFC3339),
		}
		jsonReply, _ := json.MarshalIndent(m, "", " ")
		io.WriteString(w, string(jsonReply))
	} else {
		errorMessage := Error{
			Code: http.StatusInternalServerError,
			Message: "ERROR while retrieving Git HEAD information.",
		}
		jsonError, _ := json.MarshalIndent(errorMessage, "", " ")
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, string(jsonError))
	}
}

func PullRepository(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	go git.RequestPull()
	m := Message{
		Message: "Request to update git repository received.",
	}
	jsonMessage, _ := json.MarshalIndent(m, "", " ")
	io.WriteString(w, string(jsonMessage))
}
