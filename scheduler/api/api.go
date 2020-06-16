package api

import (
	"log"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/redhat-gpe/scheduler/api/v1"
	"io"
)

func healthHandler (w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	io.WriteString(w, "OK\n")
}

func Serve() {
	router := httprouter.New()

	// Health and status checks
	router.GET("/health", healthHandler)
	router.GET("/healthz", healthHandler)

	// v1
	router.GET("/api/v1/clouds", v1.GetClouds)
	router.GET("/api/v1/clouds/:name", v1.GetCloudByName)
	router.GET("/api/v1/repo", v1.GetRepository)
	router.PUT("/api/v1/repo", v1.PullRepository)
	router.POST("/api/v1/schedule", v1.Schedule)

	log.Fatal(http.ListenAndServe(":8080", router))
}
