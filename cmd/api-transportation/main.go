package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/iot-for-tillgenglighet/api-transportation/pkg/handler"
)

func main() {
	log.Infof("Hello")

	serviceName := "api-transportation"

	log.Infof("Starting up %s ...", serviceName)

	handler.CreateRouterAndStartServing()
}
