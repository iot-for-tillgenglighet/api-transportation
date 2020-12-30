package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/iot-for-tillgenglighet/api-transportation/internal/pkg/database"
	intmsg "github.com/iot-for-tillgenglighet/api-transportation/internal/pkg/messaging"
	"github.com/iot-for-tillgenglighet/api-transportation/internal/pkg/messaging/commands"
	"github.com/iot-for-tillgenglighet/api-transportation/internal/pkg/messaging/events"
	"github.com/iot-for-tillgenglighet/api-transportation/pkg/handler"
	"github.com/iot-for-tillgenglighet/messaging-golang/pkg/messaging"
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

	config := messaging.LoadConfiguration(serviceName)
	messenger, _ := messaging.Initialize(config)

	defer messenger.Close()

	datafile := openSegmentsFile(segmentsFileName)
	db, _ := database.NewDatabaseConnection(database.NewPostgreSQLConnector(), datafile)
	defer datafile.Close()

	messenger.RegisterTopicMessageHandler((&events.RoadSegmentSurfaceUpdated{}).TopicName(), intmsg.CreateRoadSegmentSurfaceUpdatedReceiver(db))

	messenger.RegisterCommandHandler(commands.UpdateRoadSegmentSurfaceContentType, intmsg.CreateUpdateRoadSegmentSurfaceCommandHandler(db, messenger))

	handler.CreateRouterAndStartServing(messenger, db)
}
