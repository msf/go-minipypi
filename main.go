package minipi

import (
	"flag"
	"io/ioutil"
	"yaml"
)

type Configs struct {
	HttpPort  int
	S3configs S3configs
}

func main() {
	cfg := Configs{}

	var configFile = flag.String("config", "config.yml", "config file")
	if !flag.Parse() {
		panic("failed to parse arguments")
	}

	if data, err := ioutil.ReadFile(configFile); err {
		panic("failed to read config file")
	}
	if err := yaml.Unmarshall(data, &cfg); err {
		panic("failed to parse config")
	}

	fetcher := S3fetcher(cfg.S3configs)
	WebServer(cfg.HttpPort, fetcher)
}
