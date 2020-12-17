package messaging

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/iot-for-tillgenglighet/api-transportation/internal/pkg/database"
	"github.com/iot-for-tillgenglighet/api-transportation/internal/pkg/messaging/events"
	"github.com/iot-for-tillgenglighet/messaging-golang/pkg/messaging"
	"github.com/streadway/amqp"
)

//CreateRoadSegmentSurfaceUpdatedReceiver is a closure that take a datastore and handles incoming events
func CreateRoadSegmentSurfaceUpdatedReceiver(db database.Datastore) messaging.TopicMessageHandler {
	return func(msg amqp.Delivery) {
		log.Info("Message received from topic: " + string(msg.Body))

		evt := &events.RoadSegmentSurfaceUpdated{}
		err := json.Unmarshal(msg.Body, evt)

		if err != nil {
			log.Error("Failed to unmarshal message")
			return
		}

		ts, err := time.Parse(time.RFC3339, evt.Timestamp)
		err = db.UpdateRoadSegmentSurface(evt.ID, evt.SurfaceType, evt.Probability, ts)

		if err != nil {
			log.Error(err.Error())
			return
		}
	}
}
