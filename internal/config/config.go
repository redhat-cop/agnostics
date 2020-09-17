package config

import(
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/redhat-gpe/agnostics/internal/log"
	"github.com/redhat-gpe/agnostics/internal/git"
	"github.com/redhat-gpe/agnostics/internal/api/v1"
	"path/filepath"
	"path"
	"os"
)

func loadClouds() map[string]v1.Cloud {
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

	clouds := make(map[string]v1.Cloud)

	for _, cloudFile := range(cloudFileList) {
		content, err := ioutil.ReadFile(cloudFile)
		if err != nil {
			log.Err.Println("Error in loadClouds()")
			log.Err.Fatal(err)
		}
		cloud := v1.Cloud{Enabled: true}
		err = yaml.Unmarshal(content, &cloud)
		if err != nil {
			log.Err.Println("Cannot read configuration of clouds.yml")
			log.Err.Fatalf("Cannot unmarshal data: %v", err)
		} else {
			log.Debug.Printf("Found cloud %s (enabled=%v)\n", cloud.Name, cloud.Enabled)
			clouds[cloud.Name] = cloud
		}
	}

	return clouds
}

var clouds map[string]v1.Cloud

// Public functions

// Read the config from the local files and save in-memory
func Load() {
	// TODO: save and restore taints
	clouds = loadClouds()
}

// GetClouds Returns the in-memory list of clouds (v1)
func GetClouds() map[string]v1.Cloud {
	return clouds
}
