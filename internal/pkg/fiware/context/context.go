package context

import (
	"errors"
	"strings"

	"github.com/iot-for-tillgenglighet/api-transportation/internal/pkg/database"
	"github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/datamodels/fiware"
	ngsi "github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/ngsi-ld"

	log "github.com/sirupsen/logrus"
)

type contextSource struct {
	db           database.Datastore
	roads        []fiware.Road
	roadSegments []fiware.RoadSegment
}

//CreateSource instantiates and returns a Fiware ContextSource that wraps the provided db interface
func CreateSource(db database.Datastore) ngsi.ContextSource {
	return &contextSource{db: db}
}

func (cs *contextSource) CreateEntity(typeName, entityID string, req ngsi.Request) error {

	var err error

	if typeName == "Road" {
		road := &fiware.Road{}

		err := req.DecodeBodyInto(road)

		if err == nil {
			cs.roads = append(cs.roads, *road)
		}
	} else if typeName == "RoadSegment" {
		roadSegment := &fiware.RoadSegment{}

		err := req.DecodeBodyInto(roadSegment)

		if err == nil {
			cs.roadSegments = append(cs.roadSegments, *roadSegment)
		}
	}

	return err
}

func (cs *contextSource) getRoadSegments(query ngsi.Query, callback ngsi.QueryEntitiesCallback) error {
	var err error

	segments := []database.RoadSegment{}

	if query.IsGeoQuery() {
		geoQ := query.Geo()
		if geoQ.GeoRel == ngsi.GeoSpatialRelationNearPoint {
			lon, lat, _ := geoQ.Point()
			distance, _ := geoQ.Distance()
			segments, err = cs.db.GetSegmentsNearPoint(lat, lon, uint64(distance))
		} else if geoQ.GeoRel == ngsi.GeoSpatialRelationWithinRect {
			lon0, lat0, lon1, lat1, err := geoQ.Rectangle()
			if err != nil {
				return err
			}
			segments, err = cs.db.GetSegmentsWithinRect(lat0, lon0, lat1, lon1)
		}
	}

	numberOfSegments := uint64(len(segments))

	firstIndex := query.PaginationOffset()
	stopIndex := firstIndex + query.PaginationLimit()

	if stopIndex > numberOfSegments {
		stopIndex = numberOfSegments
	}

	if firstIndex > 0 || stopIndex != numberOfSegments {
		log.Infof("Returning segment %d to %d of %d", firstIndex, stopIndex-1, numberOfSegments)
	}

	for i := firstIndex; i < stopIndex; i++ {
		s := segments[i]
		rs := fiware.NewRoadSegment(s.ID(), s.ID(), s.RoadID(), s.Coordinates())
		err = callback(rs)
		if err != nil {
			break
		}
	}

	return err
}

func (cs *contextSource) GetEntities(query ngsi.Query, callback ngsi.QueryEntitiesCallback) error {

	var err error

	if query == nil {
		return errors.New("GetEntities: query may not be nil")
	}

	for _, typeName := range query.EntityTypes() {
		if typeName == "Road" {
			for _, road := range cs.roads {
				err = callback(road)
				if err != nil {
					break
				}
			}
		} else if typeName == "RoadSegment" {
			return cs.getRoadSegments(query, callback)
		}
	}

	return err
}

func (cs contextSource) ProvidesAttribute(attributeName string) bool {
	return true
}

func (cs contextSource) ProvidesEntitiesWithMatchingID(entityID string) bool {
	return strings.HasPrefix(entityID, "urn:ngsi-ld:Road:") ||
		strings.HasPrefix(entityID, "urn:ngsi-ld:RoadSegment:")
}

func (cs contextSource) ProvidesType(typeName string) bool {
	return typeName == "Road" || typeName == "RoadSegment"
}

func (cs contextSource) UpdateEntityAttributes(entityID string, req ngsi.Request) error {
	return errors.New("UpdateEntityAttributes is not supported by this service")
}
