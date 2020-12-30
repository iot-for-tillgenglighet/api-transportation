package commands

const (
	//UpdateRoadSegmentSurfaceContentType is the content type for ...
	UpdateRoadSegmentSurfaceContentType = "application/vnd-diwise-updateroadsegmentsurface+json"
)

//UpdateRoadSegmentSurface is a command that takes info about a road surface update and enqueues it for persistence
type UpdateRoadSegmentSurface struct {
	ID          string  `json:"id"`
	SurfaceType string  `json:"surfaceType"`
	Probability float64 `json:"probability"`
	Timestamp   string  `json:"timestamp"`
}

//ContentType returns the content type that this event will be sent as
func (rssu *UpdateRoadSegmentSurface) ContentType() string {
	return UpdateRoadSegmentSurfaceContentType
}
