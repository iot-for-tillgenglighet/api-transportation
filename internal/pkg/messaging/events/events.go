package events

//RoadSegmentSurfaceUpdated is an event that notifies that a road surface type has changed
type RoadSegmentSurfaceUpdated struct {
	ID          string  `json:"id"`
	SurfaceType string  `json:"surfaceType"`
	Probability float64 `json:"probability"`
	Timestamp   string  `json:"timestamp"`
}

//TopicName returns the name of the topic that this event should be posted to
func (rssu *RoadSegmentSurfaceUpdated) TopicName() string {
	return "events.transportation.roadsegmentsurfaceupdated"
}

//ContentType returns the content type that this event will be sent as
func (rssu *RoadSegmentSurfaceUpdated) ContentType() string {
	return "application/json"
}
