package console

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	"github.com/julienschmidt/httprouter"
	"github.com/redhat-gpe/agnostics/internal/api/v1"
	"github.com/redhat-gpe/agnostics/internal/config"
	"github.com/redhat-gpe/agnostics/internal/git"
	"github.com/redhat-gpe/agnostics/internal/log"
	"github.com/redhat-gpe/agnostics/internal/placement"
	"github.com/redhat-gpe/agnostics/internal/watcher"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
)

func marshal(data interface{}) string {
	json, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Err.Fatal(err)
	}
	return string(json)
}

func toYaml(data interface{}) string {
	result, err := yaml.Marshal(data)
	if err != nil {
		log.Err.Fatal(err)
	}
	return string(result)
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
		"toYaml": toYaml,
	}

	clouds := config.GetClouds()

	commitInfo, _ := v1.NewGitCommit(git.GetRepo())

	type HomeData struct {
		Clouds map[string]v1.Cloud
		Placements []v1.Placement
		GitCommit v1.GitCommit
	}

	t := template.Must(
		template.New("layout.tmpl").Funcs(fm).ParseGlob(
			filepath.Join(templateDir,"/*.tmpl")))

	t.ExecuteTemplate(w, "layout.tmpl", HomeData {
		clouds,
		placements,
		commitInfo,
	})
}

func getReloadConfig(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "text/plain")
	go watcher.RequestPull()
	io.WriteString(w, "Request to update git repository received.\n")
}

func getConfig(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "text/plain")
	commitInfo, _ := v1.NewGitCommit(git.GetRepo())
	io.WriteString(w, toYaml(commitInfo))
}

var templateDir string

// Serve function is
func Serve(t string, addr string) {
	templateDir = t
	router := httprouter.New()

	// Protected
	router.GET("/", getDashboard)
	router.GET("/get_config", getConfig)
	router.GET("/reload_config", getReloadConfig)

	log.Out.Println("Console listen on port", addr)
	log.Err.Fatal(http.ListenAndServe(addr, router))
}
