package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/iot-for-tillgenglighet/api-transportation/internal/pkg/database"
	"github.com/iot-for-tillgenglighet/api-transportation/internal/pkg/messaging/commands"
	"github.com/iot-for-tillgenglighet/api-transportation/internal/pkg/messaging/events"
	"github.com/iot-for-tillgenglighet/messaging-golang/pkg/messaging"
	"github.com/streadway/amqp"
)

//MessagingContext is an interface that allows mocking of messaging.Context parameters
type MessagingContext interface {
	PublishOnTopic(message messaging.TopicMessage) error
	NoteToSelf(message messaging.CommandMessage) error
}

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
		err = db.RoadSegmentSurfaceUpdated(evt.ID, evt.SurfaceType, evt.Probability, ts)

		if err != nil {
			log.Error(err.Error())
			return
		}
	}
}

//CreateUpdateRoadSegmentSurfaceCommandHandler returns a handler for commands
func CreateUpdateRoadSegmentSurfaceCommandHandler(db database.Datastore, msg MessagingContext) messaging.CommandHandler {
	return func(wrapper messaging.CommandMessageWrapper) error {
		cmd := &commands.UpdateRoadSegmentSurface{}
		err := json.Unmarshal(wrapper.Body(), cmd)
		if err != nil {
			return fmt.Errorf("Failed to unmarshal command! %s", err.Error())
		}

		ts, err := time.Parse(time.RFC3339, cmd.Timestamp)
		err = db.UpdateRoadSegmentSurface(cmd.ID, cmd.SurfaceType, cmd.Probability, ts)

		//Post an event stating that a roadsegment's surface has been updated
		event := &events.RoadSegmentSurfaceUpdated{
			ID:          cmd.ID,
			SurfaceType: cmd.SurfaceType,
			Probability: cmd.Probability,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}
		msg.PublishOnTopic(event)

		return nil
	}
}
