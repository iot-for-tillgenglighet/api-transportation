package context

import (
	"errors"
	"strings"

	"github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/datamodels/fiware"
	ngsi "github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/ngsi-ld"
)

type contextSource struct {
	roads        []fiware.Road
	roadSegments []fiware.RoadSegment
}

//CreateSource instantiates and returns a Fiware ContextSource that wraps the provided db interface
func CreateSource() ngsi.ContextSource {
	return &contextSource{}
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
			for _, roadSegment := range cs.roadSegments {
				err = callback(roadSegment)
				if err != nil {
					break
				}
			}
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
