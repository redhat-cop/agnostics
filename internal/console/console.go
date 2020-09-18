package console

import (
	"encoding/json"
	"html/template"
	"github.com/julienschmidt/httprouter"
	"github.com/redhat-gpe/agnostics/internal/config"
	"github.com/redhat-gpe/agnostics/internal/api/v1"
	"github.com/redhat-gpe/agnostics/internal/log"
	"github.com/redhat-gpe/agnostics/internal/placement"
	"io"
	"path/filepath"
	"net/http"
)

func marshal(data interface{}) string {
	json, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Err.Fatal(err)
	}
	return string(json)
}

func countPlacements(name string) string{
	reply, err := placement.GetCountPlacementsByCloud(name)
	if err != nil {
		return "ERROR"
	}
	return reply
}

func getDashboard(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	w.Header().Set("Content-Type", "text/html")
	placements, err := placement.GetAll(100)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "ERROR")
		log.Err.Println(err)
		return
	}

	var fm = template.FuncMap{
		"marshal": marshal,
		"countPlacements": countPlacements,
	}

	clouds := config.GetClouds()

	type HomeData struct {
		Clouds map[string]v1.Cloud
		Placements []v1.Placement
	}

	t := template.Must(
		template.New("layout.tmpl").Funcs(fm).ParseGlob(
			filepath.Join(templateDir,"/*.tmpl")))

	t.ExecuteTemplate(w, "layout.tmpl", HomeData {
		clouds,
		placements,
	})
}

var templateDir string

// Serve function is
func Serve(t string) {
	templateDir = t
	router := httprouter.New()

	// Protected
	router.GET("/", getDashboard)

	log.Out.Println("Console listen on port :8081")
	log.Err.Fatal(http.ListenAndServe(":8081", router))
}
