package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	"github.com/golang/glog"
	"github.com/konkers/lacodex"
)

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	configData, err := ioutil.ReadFile("config.json")
	if err != nil {
		glog.Fatalf("Can't open config.json: %v", err)
	}

	var config lacodex.Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		glog.Fatalf("Can't decode config.json: %v", err)
	}

	l, err := lacodex.NewLaCodex(&config)
	if err != nil {
		glog.Fatalf("Run error %v", err)
	}

	l.Run()
}
