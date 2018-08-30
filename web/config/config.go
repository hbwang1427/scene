package config

import (
	"io/ioutil"
	"log"

	"github.com/BurntSushi/toml"
)

var (
	cfgPath string = "web.toml"
	cfg     *Config
)

type Config struct {
	Http struct {
		Bind      string
		AssetsDir string
		UploadDir string
		SecretKey string
	}

	Grpc struct {
		Addr string
		Cert string
		Host string
	}

	Options struct {
		WeatherApiKey   string
		GoogleMapApiKey string
		GoogleMapDomain string
	}

	Db struct {
		Host     string
		Port     int
		User     string
		Password string
		DbName   string
	}
}

func SetConfigPath(p string) {
	log.Printf("set config path:%s", p)
	cfgPath = p
}

func parseConfig(conf string) (*Config, error) {
	var c Config
	content, err := ioutil.ReadFile(conf)
	if err != nil {
		return nil, err
	}
	if _, err = toml.Decode(string(content), &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func GetConfig() *Config {
	if cfg != nil {
		return cfg
	}

	//parse config
	var c Config
	content, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		log.Fatalf("read config file %s error:%v", cfgPath, err)
	}
	if _, err = toml.Decode(string(content), &c); err != nil {
		log.Fatalf("decode config file error:%v", err)
	}
	cfg = &c
	return &c
}
