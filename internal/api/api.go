package api

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/redhat-gpe/agnostics/internal/db"
	"github.com/redhat-gpe/agnostics/internal/log"
	"io"
	"strings"
	"encoding/base64"
	"bytes"
	"github.com/tg123/go-htpasswd"
	"path/filepath"
)

func healthHandler (w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	conn, err := db.Dial()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "ERROR: can't connect to redis\n")
		return
	}
	defer conn.Close()

	io.WriteString(w, "OK\n")
}

func BasicAuth(h httprouter.Handle, myauth *htpasswd.File, authEnabled bool) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		if ! authEnabled {
			// Delegate request to the given handle
			h(w, r, ps)
			return
		}

		const basicAuthPrefix string = "Basic "

		// Get the Basic Authentication credentials
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, basicAuthPrefix) {
			// Check credentials
			payload, err := base64.StdEncoding.DecodeString(auth[len(basicAuthPrefix):])
			if err == nil {
				pair := bytes.SplitN(payload, []byte(":"), 2)
				if len(pair) == 2 && myauth.Match(string(pair[0]), string(pair[1])) {
					// Delegate request to the given handle
					h(w, r, ps)
					return
				}
			}
		}

		// Request Basic Authentication otherwise
		w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}
}

func Serve(addr string, apiAuth bool, apiHtpasswd string) {
	router := httprouter.New()

	// Health and status checks
	router.GET("/health", healthHandler)
	router.GET("/healthz", healthHandler)
	router.GET("/api/v1/health", healthHandler)
	router.GET("/api/v1/healthz", healthHandler)

	// htpasswd authentication
	if ! apiAuth {
		apiHtpasswd = "/dev/null"
		log.Out.Println("API authentication disabled")
	}
	absAPIHtpasswdPath, err := filepath.Abs(apiHtpasswd)
	if err != nil {
		log.Err.Println("ERROR with api-htpasswd-path")
		log.Err.Fatal(err)
	}
	myauth, err := htpasswd.New(absAPIHtpasswdPath, htpasswd.DefaultSystems, nil)
	if err != nil {
		log.Err.Println("ERROR loading htpasswd", absAPIHtpasswdPath)
		log.Err.Fatal(err)
	} else {
		if apiAuth {
			log.Out.Println("htpasswd found:", absAPIHtpasswdPath)
		}
	}

	// v1
	router.GET("/api/v1/clouds", BasicAuth(v1GetClouds, myauth, apiAuth))
	router.GET("/api/v1/clouds/:name", BasicAuth(v1GetCloudByName, myauth, apiAuth))
	router.POST("/api/v1/taint/:cloudname", BasicAuth(v1PostTaintByCloudName, myauth, apiAuth))
	router.POST("/api/v1/taint/:cloudname/delete", BasicAuth(v1DeleteTaintByCloudName, myauth, apiAuth))
	router.DELETE("/api/v1/taint/:cloudname/:taintindex", BasicAuth(v1DeleteTaintByIndex, myauth, apiAuth))
	router.DELETE("/api/v1/taints/:cloudname", BasicAuth(v1DeleteTaintsByCloudName, myauth, apiAuth))
	router.GET("/api/v1/repo", BasicAuth(v1GetRepository, myauth, apiAuth))
	router.PUT("/api/v1/repo", BasicAuth(v1PullRepository, myauth, apiAuth))
	router.POST("/api/v1/schedule", BasicAuth(v1PostSchedule, myauth, apiAuth))
	router.GET("/api/v1/placements", BasicAuth(v1GetPlacements, myauth, apiAuth))
	router.GET("/api/v1/placements/:uuid", BasicAuth(v1GetPlacement, myauth, apiAuth))
	router.DELETE("/api/v1/placements/:uuid", BasicAuth(v1DeletePlacement, myauth, apiAuth))
	router.PUT("/api/v1/counters", BasicAuth(v1PutCounters, myauth, apiAuth))

	log.Out.Println("API listen on port", addr)
	log.Err.Fatal(http.ListenAndServe(addr, router))
}
