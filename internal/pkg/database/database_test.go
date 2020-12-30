package database_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	db "github.com/iot-for-tillgenglighet/api-transportation/internal/pkg/database"
	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log.SetFormatter(&log.JSONFormatter{})
	os.Exit(m.Run())
}

func TestSeedSingleRoad(t *testing.T) {
	seedData := "21277:153930;21277:153930;62.389109;17.310863;62.389084;17.310852\n"

	db, _ := db.NewDatabaseConnection(db.NewSQLiteConnector(), strings.NewReader(seedData))

	if db.GetRoadCount() != 1 {
		t.Error("Unexpected number of roads in datastore after test.", 1, "!=", db.GetRoadCount())
	}
}

func TestSeedDatabase(t *testing.T) {
	seedData := "21277:153930;21277:153930;62.389109;17.310863;62.389084;17.310852;62.389073;17.310854;62.389059;17.310878;62.389057;17.310897;62.389052;17.310940\n"

	db, _ := db.NewDatabaseConnection(db.NewSQLiteConnector(), strings.NewReader(seedData))

	_, err := db.GetRoadByID("21277:153930")
	if err != nil {
		t.Error("Unable to find expected road from id:", err.Error())
	}
}

func TestGetRoadSegmentNearPoint(t *testing.T) {
	seedData := "21277:153930;21277:153930;62.389109;17.310863;62.389084;17.310852;62.389073;17.310854;62.389059;17.310878;62.389057;17.310897;62.389052;17.310940\n"

	datastore, _ := db.NewDatabaseConnection(db.NewSQLiteConnector(), strings.NewReader(seedData))

	segments, _ := datastore.GetSegmentsNearPoint(62.389077, 17.310243, 75)
	if len(segments) == 0 {
		t.Error("Unable to find segments near a point. None returned.")
	}
}

func TestGetRoadSegmentsWithinRect(t *testing.T) {
	seedData := "21277:153930;21277:153930;62.389109;17.310863;62.389084;17.310852;62.389073;17.310854;62.389059;17.310878;62.389057;17.310897;62.389052;17.310940\n"

	datastore, _ := db.NewDatabaseConnection(db.NewSQLiteConnector(), strings.NewReader(seedData))

	segments, _ := datastore.GetSegmentsWithinRect(62.389077, 17.310243, 62.4, 17.4)
	if len(segments) == 0 {
		t.Error("Unable to find segments near a point. None returned.")
	}
}

func TestBoundingBoxCreation(t *testing.T) {
	r1 := db.NewRectangle(db.NewPoint(1, 1), db.NewPoint(2, 2))
	r2 := db.NewRectangle(db.NewPoint(1, 3), db.NewPoint(2, 4))
	r3 := db.NewRectangle(db.NewPoint(3, 1), db.NewPoint(4, 2))
	r4 := db.NewRectangle(db.NewPoint(3, 3), db.NewPoint(4, 4))

	box := db.NewBoundingBoxFromRectangles(r1, r2)
	if db.NewPoint(1.5, 2.5).IsBoundedBy(&box) == false {
		t.Error("Failed!")
	}

	box = db.NewBoundingBoxFromRectangles(r1, r3)
	if db.NewPoint(2, 1.5).IsBoundedBy(&box) == false {
		t.Error("Failed!")
	}

	box = db.NewBoundingBoxFromRectangles(r4, r1)
	if db.NewPoint(2.5, 2.5).IsBoundedBy(&box) == false {
		t.Error("Failed!")
	}
}

func TestUpdateRoadSegmentSurface(t *testing.T) {
	segmentID := "21277:153930"
	seedData := fmt.Sprintf("%s;%s;62.389109;17.310863;62.389084;17.310852\n", segmentID, segmentID)
	db, _ := db.NewDatabaseConnection(db.NewSQLiteConnector(), strings.NewReader(seedData))

	db.RoadSegmentSurfaceUpdated(segmentID, "snow", 75.0, time.Now())

	seg, _ := db.GetRoadSegmentByID(segmentID)
	surfaceType, probability := seg.SurfaceType()

	if surfaceType != "snow" || probability != 75.0 || seg.DateModified() == nil {
		t.Errorf("Failed to update road segment surface type. %s (%f) did not match expectations.", surfaceType, probability)
	}
}

func TestConnectToSQLite(t *testing.T) {
	segmentID := "21277:153930"
	seedData := fmt.Sprintf("%s;%s;62.389109;17.310863;62.389084;17.310852\n", segmentID, segmentID)
	db, _ := db.NewDatabaseConnection(db.NewSQLiteConnector(), strings.NewReader(seedData))

	err := db.UpdateRoadSegmentSurface(segmentID, "snow", 75.0, time.Now())

	if err != nil {
		t.Errorf("Failed to update road segment surface type in database. %s", err.Error())
	}

	err = db.UpdateRoadSegmentSurface(segmentID, "tarmac", 85.0, time.Now())

	if err != nil {
		t.Errorf("Failed to update road segment surface type a second time in database. %s", err.Error())
	}
}
