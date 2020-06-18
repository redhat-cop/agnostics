package config

import(
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/redhat-gpe/scheduler/log"
	"github.com/redhat-gpe/scheduler/git"
	"path/filepath"
	"path"
	"os"
)

// The Cloud type defines a cloud
type Cloud struct {
	Name string `json:"name"`
	Labels map[string]string `json:"labels"`
	// Weight is usually not provided by config, but automatically filled
	// by the scheduler later, depending on priorities configured.
	// It's possible to add it to the config though, if needed.
	Weight int `json:"weight"`
	// Enabled defines if the cloud can be selected when loading the config. It's a top-level control. If it's set to false, then the cloud will not be loaded in the configuration. It takes precedence over scheduling, thus over taints and tolerations.
	// True by default.
	Enabled bool `json:"enabled"`
}

func loadClouds() map[string]Cloud {
	cloudFileList := []string{}
	log.Debug.Println(filepath.Join(git.GetRepoDir(), "/clouds"))
	err := filepath.Walk(filepath.Join(git.GetRepoDir(), "/clouds"),
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				log.Err.Printf("%q: %v\n", p, err)
				return err
			}

			switch path.Ext(info.Name()) {
			case ".yml", ".yaml":
				cloudFileList = append(cloudFileList, p)
			}
			return nil
		})

	if err != nil {
		log.Err.Fatal(err)
	}
	log.Debug.Printf("Found %d configuration files for clouds\n",  len(cloudFileList))

	clouds := make(map[string]Cloud)

	for _, cloudFile := range(cloudFileList) {
		content, err := ioutil.ReadFile(cloudFile)
		if err != nil {
			log.Err.Println("Error in loadClouds()")
			log.Err.Fatal(err)
		}
		cloud := Cloud{Enabled: true}
		err = yaml.Unmarshal(content, &cloud)
		if err != nil {
			log.Err.Println("Cannot read configuration of clouds.yml")
			log.Err.Fatalf("Cannot unmarshal data: %v", err)
		} else {
			log.Debug.Printf("Found cloud %s (enabled=%v)\n", cloud.Name, cloud.Enabled)
			if cloud.Enabled {
				clouds[cloud.Name] = cloud
			}
		}
	}

	return clouds
}

var clouds map[string]Cloud


// Public functions

func Load() {
	clouds = loadClouds()
}

func GetClouds() map[string]Cloud {
	return clouds
}
