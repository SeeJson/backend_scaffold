package main

import (
	"flag"

	"github.com/SeeJson/backend_scaffold/app/backend"
	"github.com/SeeJson/backend_scaffold/utils"
	"github.com/SeeJson/backend_scaffold/version"
)

func main() {
	version.HandleVersion()

	var (
		addr     string
		confPath string
	)
	flag.StringVar(&confPath, "conf", "etc/conf.yml", "configuration file")
	flag.StringVar(&addr, "port", ":18001", "http addr to listen")
	flag.Parse()

	conf := new(backend.Config)
	utils.Panic(utils.LoadConfigByViper(confPath, conf))

	s, err := conf.New()
	utils.Panic(err)

	utils.Panic(s.Run(addr))
}
