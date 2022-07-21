package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Endpoint []Endpoint
}

type Endpoint struct {
	Path    string
	Service string
}

var (
	config            Config
	pathServiceTarget = map[string]string{}
	// fallbackHost      string
)

func init() {
	configFile := "config.toml"

	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatalln(err.Error())
	}

	for _, endpoint := range config.Endpoint {
		pathServiceTarget[endpoint.Path] = endpoint.Service
	}
}

func main() {
	address := viper.GetString("server.address")

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   "localhost",
	})
	proxy.Director = modifyRequest
	http.Handle("/", proxy)
	log.Info("service is running on port " + address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatal(err)
	}
}

func modifyRequest(r *http.Request) {
	pathSplitted := strings.Split(r.URL.Path, "/")

	firstPath := ""
	if len(pathSplitted) > 1 && pathSplitted[1] != "" {
		firstPath = pathSplitted[1]
	}

	host := ""
	if pathServiceTarget[firstPath] != "" {
		host = pathServiceTarget[firstPath]
	}
	target, _ := url.Parse(host)

	r.Host = target.Host
	r.URL.Host = r.Host
	r.URL.Scheme = target.Scheme
}
