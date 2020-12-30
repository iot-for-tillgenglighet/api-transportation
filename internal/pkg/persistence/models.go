package persistence

import (
	"time"

	"gorm.io/gorm"
)

//Road persists the bare minimum we need to store about a road
type Road struct {
	gorm.Model
	RID          string `gorm:"unique"`
	RoadSegments []RoadSegment
}

//RoadSegment persists the bare minimum we need to store in a database about a road segment
type RoadSegment struct {
	gorm.Model
	SegmentID              string `gorm:"unique"`
	RoadID                 uint
	SurfaceTypePredictions []SurfaceTypePrediction
}

//SurfaceTypePrediction is a model for a temporary table until a better schema is designed
type SurfaceTypePrediction struct {
	gorm.Model
	RoadSegmentID uint
	SurfaceType   string
	Probability   float64
	Timestamp     time.Time
}
