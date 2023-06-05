package main

import (
	"github.com/SeeJson/account/cmd/account/config"
	httpserver "github.com/SeeJson/account/cmd/account/server/http"
	log "github.com/sirupsen/logrus"
)

// @title Go account API
// @version 1.0
// @description account服务
func main() {
	var err error

	// load config
	err = config.Load("../../conf", "config", "yaml")
	if err != nil {
		log.Fatalf("fail to load config: %v", err)
	}

	// todo start rpc server

	// start http server
	err = httpserver.Run()
	if err != nil {
		log.Fatal(err)
	}

	// signal?
}
