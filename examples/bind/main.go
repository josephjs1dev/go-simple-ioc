package main

import (
	"log"

	"github.com/josephsalimin/go-simple-ioc/ioc"
)

type Config struct {
	isDebug bool
}

func main() {
	var boundCfg = &Config{isDebug: true}
	ioc.MustBind(&boundCfg)
	ioc.MustBindWithAlias(&boundCfg, "your_alias")

	var resolvedCfg *Config
	if err := ioc.Resolve(&resolvedCfg); err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v, %+v\n", boundCfg, resolvedCfg)
}
