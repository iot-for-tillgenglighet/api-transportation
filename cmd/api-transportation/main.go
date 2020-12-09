package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/iot-for-tillgenglighet/api-transportation/internal/pkg/database"
	"github.com/iot-for-tillgenglighet/api-transportation/pkg/handler"
)

func openSegmentsFile(path string) *os.File {
	datafile, err := os.Open(path)
	if err != nil {
		log.Infof("Failed to open the segments database file %s. Datastore will not be seeded.", path)
		return nil
	}
	return datafile
}

var segmentsFileName string

func main() {
	flag.StringVar(&segmentsFileName, "segsfile", "", "The file to seed road segments from")
	flag.Parse()

	log.SetFormatter(&log.JSONFormatter{})

	serviceName := "api-transportation"

	log.Infof("Starting up %s ...", serviceName)

	datafile := openSegmentsFile(segmentsFileName)
	db, _ := database.NewDatabaseConnection(datafile)
	defer datafile.Close()

	handler.CreateRouterAndStartServing(db)
}
