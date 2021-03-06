package context

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iot-for-tillgenglighet/api-transportation/internal/pkg/database"
	"github.com/iot-for-tillgenglighet/api-transportation/internal/pkg/messaging"
	"github.com/iot-for-tillgenglighet/api-transportation/internal/pkg/messaging/commands"
	diwise "github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/datamodels/diwise"
	"github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/datamodels/fiware"
	ngsi "github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/ngsi-ld"
	ngsitypes "github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/ngsi-ld/types"

	log "github.com/sirupsen/logrus"
)

type contextSource struct {
	db  database.Datastore
	msg messaging.MessagingContext
}

//CreateSource instantiates and returns a Fiware ContextSource that wraps the provided db interface
func CreateSource(db database.Datastore, msg messaging.MessagingContext) ngsi.ContextSource {
	return &contextSource{db: db, msg: msg}
}

func (cs *contextSource) CreateEntity(typeName, entityID string, req ngsi.Request) error {
	var err error

	if typeName == "RoadSurfaceObserved" {
		rso := &diwise.RoadSurfaceObserved{}
		err = req.DecodeBodyInto(rso)
		if err != nil {
			log.Errorf("Could not create new RoadSurfaceObserved: " + err.Error())
			return err
		}
		rso.ID = uuid.New().String()
		_, err = cs.db.CreateRoadSurfaceObserved(rso)
	}

	return err
}

func (cs *contextSource) getRoads(query ngsi.Query, callback ngsi.QueryEntitiesCallback) error {
	var err error

	roads := []database.Road{}

	if query.IsGeoQuery() {
		geoQ := query.Geo()
		if geoQ.GeoRel == ngsi.GeoSpatialRelationNearPoint {
			lon, lat, _ := geoQ.Point()
			distance, _ := geoQ.Distance()
			roads, err = cs.db.GetRoadsNearPoint(lat, lon, uint64(distance))
		} else if geoQ.GeoRel == ngsi.GeoSpatialRelationWithinRect {
			lon0, lat0, lon1, lat1, err := geoQ.Rectangle()
			if err != nil {
				return err
			}
			roads, _ = cs.db.GetRoadsWithinRect(lat0, lon0, lat1, lon1)
		}
	}

	numberOfRoads := uint64(len(roads))

	firstIndex := query.PaginationOffset()
	stopIndex := firstIndex + query.PaginationLimit()

	if stopIndex > numberOfRoads {
		stopIndex = numberOfRoads
	}

	if firstIndex > 0 || stopIndex != numberOfRoads {
		log.Infof("Returning road %d to %d of %d", firstIndex, stopIndex-1, numberOfRoads)
	}

	for i := firstIndex; i < stopIndex; i++ {
		r := roads[i]
		fwRoad := fiware.NewRoad(r.ID(), r.ID(), "class", r.GetSegmentIdentities())

		err = callback(fwRoad)
		if err != nil {
			break
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
			segments, _ = cs.db.GetSegmentsWithinRect(lat0, lon0, lat1, lon1)
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

	sort.Slice(segments, func(i, j int) bool {
		iTime := segments[i].DateModified()
		jTime := segments[j].DateModified()

		// Sort by time first with most recently updated at the top
		// TODO: Remove this when we have working pagination as it breaks pageability of the data
		if iTime != nil {
			if jTime != nil {
				// We want the most reent to come first, so it is less if it is after ...
				return iTime.After(*jTime)
			} else {
				return true
			}
		} else if jTime != nil {
			return false
		}

		// Neither segment have been updated. Sort by segment id
		return strings.Compare(segments[i].ID(), segments[j].ID()) < 0
	})

	for i := firstIndex; i < stopIndex; i++ {
		s := segments[i]
		rs := fiware.NewRoadSegment(s.ID(), s.ID(), s.RoadID(), s.Coordinates(), s.DateModified())

		surfaceType, probability := s.SurfaceType()
		rs = rs.WithSurfaceType(surfaceType, probability)

		err = callback(rs)
		if err != nil {
			break
		}
	}

	return err
}

func (cs *contextSource) getRoadSurfaceObserved(query ngsi.Query, callback ngsi.QueryEntitiesCallback) error {
	roadSurfaces, err := cs.db.GetRoadSurfacesObserved()
	if err != nil {
		return err
	}
	for _, rso := range roadSurfaces {
		diwiseRoadSurface := diwise.NewRoadSurfaceObserved(rso.RoadSurfaceObservedID, rso.SurfaceType, rso.Probability, rso.Latitude, rso.Longitude)
		diwiseRoadSurface.DateObserved = ngsitypes.CreateDateTimeProperty(rso.Timestamp.Format(time.RFC3339))
		err = callback(diwiseRoadSurface)
		if err != nil {
			break
		}
	}

	return nil
}

func (cs *contextSource) GetEntities(query ngsi.Query, callback ngsi.QueryEntitiesCallback) error {

	var err error

	if query == nil {
		return errors.New("GetEntities: query may not be nil")
	}

	for _, typeName := range query.EntityTypes() {
		if typeName == "Road" {
			return cs.getRoads(query, callback)
		} else if typeName == "RoadSegment" {
			return cs.getRoadSegments(query, callback)
		} else if typeName == "RoadSurfaceObserved" {
			return cs.getRoadSurfaceObserved(query, callback)
		}
	}

	return err
}

func (cs contextSource) ProvidesAttribute(attributeName string) bool {
	return true
}

func (cs contextSource) ProvidesEntitiesWithMatchingID(entityID string) bool {
	return strings.HasPrefix(entityID, fiware.RoadIDPrefix) ||
		strings.HasPrefix(entityID, fiware.RoadSegmentIDPrefix) ||
		strings.HasPrefix(entityID, diwise.RoadSurfaceObservedIDPrefix)
}

func (cs contextSource) ProvidesType(typeName string) bool {
	return typeName == "Road" || typeName == "RoadSegment" || typeName == "RoadSurfaceObserved"
}

func (cs contextSource) RetrieveEntity(entityID string, request ngsi.Request) (ngsi.Entity, error) {
	return nil, nil
}

func (cs contextSource) UpdateEntityAttributes(entityID string, req ngsi.Request) error {
	if !strings.Contains(entityID, ":RoadSegment:") {
		return errors.New("UpdateEntityAttributes is only supported for RoadSegments")
	}

	updateSource := &fiware.RoadSegment{}
	err := req.DecodeBodyInto(updateSource)
	if err != nil {
		log.Errorln("Failed to decode PATCH body in UpdateEntityAttributes: " + err.Error())
		return err
	}

	if updateSource.SurfaceType == nil {
		return errors.New("UpdateEntityAttributes only supports the surfaceType property which MUST be non null")
	}

	segment, err := cs.db.GetRoadSegmentByID(entityID[24:])
	if err != nil {
		return err
	}

	//Enqueue a command to a replica of this service, to persist the road surface update
	command := &commands.UpdateRoadSegmentSurface{
		ID:          segment.ID(),
		SurfaceType: strings.ToLower(updateSource.SurfaceType.Value),
		Probability: updateSource.SurfaceType.Probability,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}
	err = cs.msg.NoteToSelf(command)
	if err != nil {
		log.Error(err.Error())
		return errors.New("failed to update entity attributes")
	}

	return nil
}
