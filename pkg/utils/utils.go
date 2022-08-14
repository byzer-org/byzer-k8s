package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/jsonq"
	"io/ioutil"
	"os"
	"strings"
)

var logger = GetLogger("byzer-k8s-deploy")

func CreateTmpFile(content string) (*os.File, error) {
	//fsys := os.DirFS(".")
	tmpfile, _ := os.CreateTemp(".", "*")
	tmpfileName := tmpfile.Name()
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		return nil, fmt.Errorf("Fail to create tmp file [%s] ", tmpfileName)
	}
	return tmpfile, nil
}

func BuildJsonQueryFromStr(jsonStr string) *jsonq.JsonQuery {
	data := map[string]interface{}{}
	dec := json.NewDecoder(strings.NewReader(jsonStr))
	dec.Decode(&data)
	jq := jsonq.NewQuery(data)
	return jq

}

// ReadConfigFile returns spark storage and mlsql config in one map
func ReadConfigFile(path string) (map[string]string, error) {
	if len(path) == 0 {
		logger.Fatalf("load engine config file from %s: %s", path, errors.New("config file is required"))
	}

	logger.Infof("Read config file %s", path)

	b, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Fatalf("load engine config file from %s: %s", path, err)
	}

	lines := strings.Split(string(b), "\n")
	conf := make(map[string]string)

	for _, line := range lines {
		cleanLine := strings.TrimSpace(line)
		if cleanLine == "" {
			continue
		}
		if strings.HasPrefix(cleanLine, "#") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		conf[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}

	return conf, nil
}


