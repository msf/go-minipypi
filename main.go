package main

import (
	"flag"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type Configs struct {
	WebConfigs WebServerConfigs
	S3configs  S3configs
}

func genConfig(filename string) {
	cfg := &Configs{
		WebConfigs: WebServerConfigs{
			BasePath:           "",
			LocalFileDirectory: "/Users/miguel/cm/configs/",
			Port:               8080,
		},
		S3configs: S3configs{
			AccessKey:  "134344565452",
			SecretKey:  "324367565w/34646763/4243647/24",
			BucketName: "pakage",
		},
	}

	d, _ := yaml.Marshal(cfg)
	ioutil.WriteFile(filename+"gen", d, 0640)
}

func main() {
	cfg := Configs{}

	var configFile = flag.String("config", "config.yml", "config file")
	flag.Parse()

	genConfig("cfg.yml")
	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		genConfig("cfg.yml")
		panic("failed to read config file")
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic("failed to parse config")
	}

	log.Printf("config: %v\n", cfg)

	fetcher := NewS3Fetcher(cfg.S3configs)
	WebServer(cfg.WebConfigs, fetcher)
}
