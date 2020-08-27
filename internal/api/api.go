package api

import (
	"log"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/redhat-gpe/agnostics/internal/db"
	"io"
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

func Serve() {
	router := httprouter.New()

	// Health and status checks
	router.GET("/health", healthHandler)
	router.GET("/healthz", healthHandler)
	router.GET("/api/v1/health", healthHandler)
	router.GET("/api/v1/healthz", healthHandler)

	// v1
	router.GET("/api/v1/clouds", v1GetClouds)
	router.GET("/api/v1/clouds/:name", v1GetCloudByName)
	router.GET("/api/v1/repo", v1GetRepository)
	router.PUT("/api/v1/repo", v1PullRepository)
	router.POST("/api/v1/schedule", v1PostSchedule)
	router.GET("/api/v1/placements", v1GetPlacements)
	router.GET("/api/v1/placements/:uuid", v1GetPlacement)
	router.DELETE("/api/v1/placements/:uuid", v1DeletePlacement)

	log.Fatal(http.ListenAndServe(":8080", router))
}
