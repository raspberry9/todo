// Copyright 2015 koo@kormail.net. All rights reserved.

// Package main implements a authentication http server.
package main

import (
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/formatters/logstash"
	"jsproj.com/koo/server/auth/handlers"
	"jsproj.com/koo/server/auth/schema"
)

const (
	serverName     = "talkcrew-auth"
	serverVersion  = "1.0"
	configFilePath = "auth.cfg"
)

func init() {
	log.SetFormatter(&logstash.LogstashFormatter{Type: serverName})
	schema.MustInit(configFilePath)
	conf := schema.Config()
	if conf.IsTestServer() {
		log.SetOutput(os.Stdout)
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetOutput(os.Stderr)
		log.SetLevel(log.WarnLevel)
	}
}

func main() {
	log.Infof("%s version %s start", serverName, serverVersion)
	router := handlers.MustInit()
	bindAddr := schema.Config().Server.Bind
	http.Handle("/", router)
	log.Infof("listen on %s...", bindAddr)
	if err := http.ListenAndServe(bindAddr, http.DefaultServeMux); err != nil {
		log.Fatalf("can't start server. err=%s", err)
	}
	log.Info("bye.\n")
}
