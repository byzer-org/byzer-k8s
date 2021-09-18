package utils

import (
	"os"
	"testing"
)

func TestReadConfigFile(t *testing.T) {
	content := `engine.spark.confkey=confvalue
engine.storage.confkey=confvalue`
	f, e := CreateTmpFile(content)
	if e != nil {
		t.Error(e)
	}
	defer os.Remove(f.Name())

	conf, err := ReadConfigFile(f.Name())

	if err != nil {
		t.Error(e)
	}
	if conf["engine.spark.confkey"] != "confvalue" {
		t.Error("Config file should contain engine.spark.confkey=confvalue")
	}

	if conf["engine.storage.confkey"] != "confvalue" {
		t.Error("Config file should contain engine.storage.confkey=confvalue")
	}

}
