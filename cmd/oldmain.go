package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// Configs used by this package.
type Configs struct {
	WebConfigs   WebServerConfigs
	CacheConfigs CacheConfigs
	S3configs    S3configs
}

func genConfig(filename string) {
	cfg := &Configs{
		WebConfigs: WebServerConfigs{
			Host:     "localhost",
			BasePath: "/",
			Port:     8080,
		},
		CacheConfigs: CacheConfigs{
			ExpireSecs: 120,
		},
		S3configs: S3configs{
			BucketName:      "bucket-name",
			CredentialsFile: "aws_credentials.ini",
			Region:          "eu-west-1",
		},
	}

	d, _ := yaml.Marshal(cfg)
	ioutil.WriteFile(filename, d, 0640)
}

func main() {
	cfg := Configs{}

	var configFile = flag.String("config", "config.yml", "config file")

	flag.Parse()

	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		genConfig("config.yml.gen")
		println("failed to read config file, see example: config.yml.gen")
		os.Exit(1)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		println("failed to parse config")
		os.Exit(1)
	}

	log.Printf("config: %v\n", cfg)
	if !isValidConfig(cfg) {
		println("Invalid config")
		os.Exit(1)
	}

	fetcher := NewS3Fetcher(cfg.S3configs)
	cache := NewCachedFetcher(cfg.CacheConfigs, fetcher)
	RunWebServer(cfg.WebConfigs, cache)
}

func isValidConfig(config Configs) bool {
	valid := true
	valid = valid && len(config.WebConfigs.BasePath) > 0
	valid = valid && config.WebConfigs.Port > 0
	valid = valid && config.WebConfigs.Port < 65535
	valid = valid && len(config.S3configs.BucketName) > 0
	valid = valid && len(config.S3configs.CredentialsFile) > 0
	valid = valid && len(config.S3configs.Region) > 0
	return valid
}
