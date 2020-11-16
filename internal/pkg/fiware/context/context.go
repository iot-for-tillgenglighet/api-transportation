package context

import (
	"errors"
	"strings"

	"github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/datamodels/fiware"
	ngsi "github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/ngsi-ld"
)

type contextSource struct {
	roads []fiware.Road
}

//CreateSource instantiates and returns a Fiware ContextSource that wraps the provided db interface
func CreateSource() ngsi.ContextSource {
	return &contextSource{}
}

func (cs *contextSource) CreateEntity(typeName, entityID string, req ngsi.Request) error {

	road := &fiware.Road{}
	err := req.DecodeBodyInto(road)

	if err == nil {
		cs.roads = append(cs.roads, *road)
	}

	return err
}

func (cs *contextSource) GetEntities(query ngsi.Query, callback ngsi.QueryEntitiesCallback) error {

	var err error

	for _, serviceRequest := range cs.roads {
		err = callback(serviceRequest)
		if err != nil {
			break
		}
	}

	return err
}

func (cs contextSource) ProvidesAttribute(attributeName string) bool {
	return true
}

func (cs contextSource) ProvidesEntitiesWithMatchingID(entityID string) bool {
	return strings.HasPrefix(entityID, "urn:ngsi-ld:Road:")
}

func (cs contextSource) ProvidesType(typeName string) bool {
	return typeName == "Road"
}

func (cs contextSource) UpdateEntityAttributes(entityID string, req ngsi.Request) error {
	return errors.New("UpdateEntityAttributes is not supported by this service")
}
