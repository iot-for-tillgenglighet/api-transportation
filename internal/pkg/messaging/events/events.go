package events

//RoadSegmentSurfaceUpdated is an event that notifies that a road surface type has changed
type RoadSegmentSurfaceUpdated struct {
	ID          string  `json:"id"`
	SurfaceType string  `json:"surfaceType"`
	Probability float64 `json:"probability"`
}

func (rssu *RoadSegmentSurfaceUpdated) TopicName() string {
	return "events-roadsegmentsurfaceupdated"
}

func (msg *RoadSegmentSurfaceUpdated) ContentType() string {
	return "application/json"
}
