package database

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

//TODO: This is a proof of concept that is in need to be refactored to use a
// 		proper 3rd party geometry library like S2 or something like that. // Isak

//Point encapsulates a WGS84 coordinate
type Point struct {
	lat float64
	lon float64
}

//NewPoint creates a new point instance to encapsulate the provided coordinate
func NewPoint(lat, lon float64) Point {
	return Point{lat: lat, lon: lon}
}

//IsBoundedBy returns true if the point is bounded by the provided bounding box
func (p Point) IsBoundedBy(box *Rectangle) bool {
	if box.northWest.lon < p.lon && box.southEast.lon > p.lon &&
		box.northWest.lat > p.lat && box.southEast.lat < p.lat {
		//log.Infof("point (%f,%f) is bounded by (%f,%f)(%f,%f)", p.lat, p.lon, box.northWest.lat, box.northWest.lon, box.southEast.lat, box.southEast.lon)
		return true
	}

	//log.Infof("point (%f,%f) is NOT bounded by (%f,%f)(%f,%f)", p.lat, p.lon, box.northWest.lat, box.northWest.lon, box.southEast.lat, box.southEast.lon)

	return false
}

//Rectangle is a rectangle shaped box based on a NW and a SE coordinate
type Rectangle struct {
	northWest Point
	southEast Point
}

//NewRectangle takes two Points and returns a new Rectangle instance
func NewRectangle(pt1 Point, pt2 Point) Rectangle {
	nw := NewPoint(math.Max(pt1.lat, pt2.lat), math.Min(pt1.lon, pt2.lon))
	se := NewPoint(math.Min(pt1.lat, pt2.lat), math.Max(pt1.lon, pt2.lon))

	return Rectangle{northWest: nw, southEast: se}
}

//NewBoundingBoxFromRectangles creates a new instance of a Rectangle by creating a union of two others
func NewBoundingBoxFromRectangles(rect1, rect2 Rectangle) Rectangle {

	nw := NewPoint(math.Max(rect1.northWest.lat, rect2.northWest.lat), math.Min(rect1.northWest.lon, rect2.northWest.lon))
	se := NewPoint(math.Min(rect1.southEast.lat, rect2.southEast.lat), math.Max(rect1.southEast.lon, rect2.southEast.lon))

	result := NewRectangle(nw, se)

	/*log.Infof("(%f,%f)(%f,%f) unioned with (%f,%f)(%f,%f) == (%f,%f)(%f,%f)",
		rect1.northWest.lat, rect1.northWest.lon, rect1.southEast.lat, rect1.southEast.lon,
		rect2.northWest.lat, rect2.northWest.lon, rect2.southEast.lat, rect2.southEast.lon,
		result.northWest.lat, result.northWest.lon, result.southEast.lat, result.southEast.lon,
	)*/

	return result
}

//DistanceFromPoint calculates the distance from a rectangle and an exterior point. For
//points within the rectangle, 0 is returned
func (r Rectangle) DistanceFromPoint(pt Point) uint64 {
	latdelta := math.Max(math.Max(r.southEast.lat-pt.lat, 0), pt.lat-r.northWest.lat)
	londelta := math.Max(math.Max(r.northWest.lon-pt.lon, 0), pt.lon-r.southEast.lon)
	distance := 0.0

	if latdelta > 0 || londelta > 0 {
		LatB := latdelta * math.Pi / 180
		LngB := londelta * math.Pi / 180

		distance = 6371000 * math.Acos(math.Cos(LatB)*math.Cos(LngB))
	}

	return uint64(distance)
}

//Intersects returns true if the two rectangles overlap in any way
func (r Rectangle) Intersects(other Rectangle) bool {
	if r.southEast.lon <= other.northWest.lon || r.southEast.lat >= other.northWest.lat ||
		r.northWest.lon >= other.southEast.lon || r.northWest.lat <= other.southEast.lat {
		return false
	}

	return true
}

//Road is a road
type Road interface {
	ID() string

	AddSegment(RoadSegment)
	GetSegment(id string) (RoadSegment, error)
	GetSegmentIdentities() []string
	GetSegmentsWithinDistanceFromPoint(maxDistance uint64, pt Point) ([]RoadSegment, uint64)
	GetSegmentsWithinRect(Rectangle) ([]RoadSegment, uint64)

	BoundingBox() Rectangle
	IsWithinDistanceFromPoint(maxDistance uint64, pt Point) bool

	setLastModified(timestamp *time.Time)
}

type roadImpl struct {
	id string

	segments []RoadSegment

	bbox     Rectangle
	modified *time.Time
}

func (r *roadImpl) AddSegment(segment RoadSegment) {
	r.segments = append(r.segments, segment)
}

func (r *roadImpl) GetSegment(id string) (RoadSegment, error) {
	for idx := range r.segments {
		if r.segments[idx].ID() == id {
			//TODO: This should return a copy of the segment, and not an interface pointing to the actual segment
			return r.segments[idx], nil
		}
	}

	return nil, fmt.Errorf("not found")
}

func (r *roadImpl) GetSegmentIdentities() []string {
	identities := []string{}
	for idx := range r.segments {
		identities = append(identities, r.segments[idx].ID())
	}
	return identities
}

func (r *roadImpl) BoundingBox() Rectangle {
	return r.bbox
}

func (r *roadImpl) GetSegmentsWithinDistanceFromPoint(maxDistance uint64, pt Point) ([]RoadSegment, uint64) {
	matchingSegments := []RoadSegment{}
	count := uint64(0)

	for _, segment := range r.segments {
		if segment.IsWithinDistanceFromPoint(maxDistance, pt) {
			matchingSegments = append(matchingSegments, segment)
			count++
		}
	}

	return matchingSegments, count
}

func (r *roadImpl) GetSegmentsWithinRect(rect Rectangle) ([]RoadSegment, uint64) {
	matchingSegments := []RoadSegment{}
	count := uint64(0)

	for _, segment := range r.segments {
		if segment.BoundingBox().Intersects(rect) {
			matchingSegments = append(matchingSegments, segment)
			count++
		}
	}

	return matchingSegments, count
}

func (r *roadImpl) ID() string {
	return r.id
}

func (r *roadImpl) IsWithinDistanceFromPoint(maxDistance uint64, pt Point) bool {
	return (maxDistance > r.bbox.DistanceFromPoint(pt))
}

func (r *roadImpl) setLastModified(timestamp *time.Time) {
	r.modified = timestamp
}

func newRoad(id string, segment RoadSegment) Road {
	road := &roadImpl{id: id}
	road.segments = append(road.segments, segment)
	road.bbox = segment.BoundingBox()

	return road
}

//RoadSegment is a road segment
type RoadSegment interface {
	ID() string
	RoadID() string
	BoundingBox() Rectangle
	Coordinates() [][2]float64
	IsWithinDistanceFromPoint(uint64, Point) bool
	SurfaceType() (string, float64)

	setSurfaceType(surfaceType string, probability float64)

	DateModified() *time.Time
	IsModified() bool
	setLastModified(timestamp *time.Time)
}

type roadSegmentImpl struct {
	id     string
	roadID string

	lines []RoadSegmentLine
	bbox  Rectangle

	surfaceType            string
	surfaceTypeProbability float64

	modified *time.Time
}

func (seg *roadSegmentImpl) ID() string {
	return seg.id
}

func (seg *roadSegmentImpl) RoadID() string {
	return seg.roadID
}

func (seg *roadSegmentImpl) BoundingBox() Rectangle {
	return seg.bbox
}

func (seg *roadSegmentImpl) Coordinates() [][2]float64 {
	coords := [][2]float64{seg.lines[0].StartPoint()}

	for idx := range seg.lines {
		coords = append(coords, seg.lines[idx].EndPoint())
	}

	return coords
}

func (seg *roadSegmentImpl) IsWithinDistanceFromPoint(maxDistance uint64, pt Point) bool {

	for _, line := range seg.lines {
		if line.BoundingBox().DistanceFromPoint(pt) < maxDistance {
			return true
		}
	}

	return false
}

func (seg *roadSegmentImpl) SurfaceType() (string, float64) {
	return seg.surfaceType, seg.surfaceTypeProbability
}

func (seg *roadSegmentImpl) setSurfaceType(surfaceType string, probability float64) {
	seg.surfaceType = surfaceType
	seg.surfaceTypeProbability = probability
}

func (seg *roadSegmentImpl) DateModified() *time.Time {
	return seg.modified
}

func (seg *roadSegmentImpl) IsModified() bool {
	return seg.modified != nil
}

func (seg *roadSegmentImpl) setLastModified(timestamp *time.Time) {
	if seg.modified == nil || seg.modified.Before(*timestamp) {
		seg.modified = timestamp
	}
}

func newRoadSegment(id string, roadID string, coordinates []Point) RoadSegment {
	lines := []RoadSegmentLine{}
	line := newRoadSegmentLine(coordinates[0], coordinates[1])
	bbox := line.BoundingBox()

	lines = append(lines, line)

	for i := 1; i < len(coordinates)-1; i++ {
		line := newRoadSegmentLine(coordinates[i], coordinates[i+1])
		lines = append(lines, line)
		bbox = NewBoundingBoxFromRectangles(bbox, line.BoundingBox())
	}

	/*log.Infof("Created segment with bbox (%f,%f),(%f,%f).",
	bbox.northWest.lat, bbox.northWest.lon,
	bbox.southEast.lat, bbox.southEast.lon)*/

	return &roadSegmentImpl{id: id, roadID: roadID, bbox: bbox, lines: lines}
}

//RoadSegmentLine represents a straight part of a road segment
type RoadSegmentLine interface {
	BoundingBox() Rectangle
	StartPoint() [2]float64
	EndPoint() [2]float64
}

type roadSegmentLineImpl struct {
	startPt Point
	endPt   Point
	bbox    Rectangle
}

func (line roadSegmentLineImpl) BoundingBox() Rectangle {
	return line.bbox
}

func (line roadSegmentLineImpl) EndPoint() [2]float64 {
	return [2]float64{line.endPt.lon, line.endPt.lat}
}

func (line roadSegmentLineImpl) StartPoint() [2]float64 {
	return [2]float64{line.startPt.lon, line.startPt.lat}
}

func newRoadSegmentLine(startPt Point, endPt Point) RoadSegmentLine {
	bbox := NewRectangle(startPt, endPt)
	line := roadSegmentLineImpl{startPt: startPt, endPt: endPt, bbox: bbox}
	return line
}

//Datastore is an interface that is used to inject the database into different handlers to improve testability
type Datastore interface {
	AddRoad(Road) error

	GetRoadByID(id string) (Road, error)
	GetRoadCount() int
	GetRoadsNearPoint(lat, lon float64, maxDistance uint64) ([]Road, error)
	GetRoadsWithinRect(lat0, lon0, lat1, lon1 float64) ([]Road, error)

	GetRoadSegmentByID(id string) (RoadSegment, error)

	GetSegmentsNearPoint(lat, lon float64, maxDistance uint64) ([]RoadSegment, error)
	GetSegmentsWithinRect(lat0, lon0, lat1, lon1 float64) ([]RoadSegment, error)

	UpdateRoadSegmentSurface(segmentID, surfaceType string, probability float64, timestamp time.Time) error
}

//InitFromReader takes a reader interface and initialises the datastore
func initFromReader(db *myDB, rd io.Reader) error {
	// Start reading from the file with a reader.
	reader := bufio.NewReader(rd)
	var line string
	var err error

	log.Infof("Seeding datastore ...")

	for {
		line, err = reader.ReadString('\n')
		if err != nil && err != io.EOF {
			break
		}

		parts := strings.Split(line, ";")
		numberOfParts := len(parts)

		if numberOfParts >= 6 {
			coordinates := []Point{}

			parts[numberOfParts-1] = strings.TrimRight(parts[numberOfParts-1], "\r\n")

			for i := 2; i < numberOfParts; i += 2 {
				lat, laterr := strconv.ParseFloat(parts[i], 64)
				lon, lonerr := strconv.ParseFloat(parts[i+1], 64)

				if laterr != nil || lonerr != nil {
					log.Errorf("Failed to parse (%f,%f) as a coordinate. Skipping record.", lat, lon)
					continue
				}

				coordinates = append(coordinates, NewPoint(lat, lon))
			}

			segment := newRoadSegment(parts[1], parts[0], coordinates)
			road := newRoad(parts[0], segment)

			db.AddRoad(road)
		}

		if err != nil {
			break
		}
	}

	if err != io.EOF {
		log.Errorf(" > Failed with error: %v\n", err)
		return err
	}

	return nil
}

//NewDatabaseConnection creates and returns a new instance of the Datastore interface
func NewDatabaseConnection(datafile io.Reader) (Datastore, error) {
	db := &myDB{}

	if datafile != nil {
		err := initFromReader(db, datafile)
		if err != nil {
			return nil, err
		}

		log.Infof("Datastore seeded with %d roads.", db.GetRoadCount())

		db.GetRoadsWithinRect(62.430242, 17.230700, 62.353557, 17.444075)
		db.GetSegmentsWithinRect(62.430242, 17.230700, 62.353557, 17.444075)
	}

	return db, nil
}

func (db *myDB) AddRoad(road Road) error {
	db.roads = append(db.roads, road)
	return nil
}

func (db *myDB) GetRoadByID(id string) (Road, error) {
	for _, road := range db.roads {
		if road.ID() == id {
			return road, nil
		}
	}

	return nil, fmt.Errorf("No road with id %s in datastore", id)
}

func (db *myDB) GetRoadCount() int {
	return len(db.roads)
}

func (db *myDB) GetRoadsNearPoint(lat, lon float64, maxDistance uint64) ([]Road, error) {
	roads := []Road{}

	pt := NewPoint(lat, lon)

	for _, road := range db.roads {
		if road.IsWithinDistanceFromPoint(maxDistance, pt) {
			roads = append(roads, road)
		}
	}

	return roads, nil
}

func (db *myDB) GetRoadsWithinRect(lat0, lon0, lat1, lon1 float64) ([]Road, error) {
	roads := []Road{}

	rect := NewRectangle(NewPoint(lat0, lon0), NewPoint(lat1, lon1))

	for _, road := range db.roads {
		if rect.Intersects(road.BoundingBox()) {
			roads = append(roads, road)
		}
	}

	log.Infof("Found %d roads within rect (%f,%f)(%f,%f).", len(roads), rect.northWest.lat, rect.northWest.lon, rect.southEast.lat, rect.southEast.lon)

	return roads, nil
}

func (db *myDB) GetRoadSegmentByID(id string) (RoadSegment, error) {

	for idx := range db.roads {
		segment, err := db.roads[idx].GetSegment(id)
		if err == nil {
			return segment, nil
		}
	}

	return nil, fmt.Errorf("Unable to find RoadSegment with id %s", id)
}

func (db *myDB) GetSegmentsNearPoint(lat, lon float64, maxDistance uint64) ([]RoadSegment, error) {
	segments := []RoadSegment{}

	pt := NewPoint(lat, lon)

	for _, road := range db.roads {
		if road.IsWithinDistanceFromPoint(maxDistance, pt) {
			roadsegments, count := road.GetSegmentsWithinDistanceFromPoint(maxDistance, pt)
			if count > 0 {
				segments = append(segments, roadsegments...)
			}
		}
	}

	return segments, nil
}

func (db *myDB) GetSegmentsWithinRect(lat0, lon0, lat1, lon1 float64) ([]RoadSegment, error) {
	segments := []RoadSegment{}

	rect := NewRectangle(NewPoint(lat0, lon0), NewPoint(lat1, lon1))

	for _, road := range db.roads {
		if rect.Intersects(road.BoundingBox()) {
			roadsegs, count := road.GetSegmentsWithinRect(rect)
			if count > 0 {
				segments = append(segments, roadsegs...)
			}
		}
	}

	log.Infof("Found %d segments within rect (%f,%f)(%f,%f).", len(segments), rect.northWest.lat, rect.northWest.lon, rect.southEast.lat, rect.southEast.lon)

	return segments, nil
}

func (db *myDB) UpdateRoadSegmentSurface(segmentID, surfaceType string, probability float64, timestamp time.Time) error {

	for idx := range db.roads {
		segment, err := db.roads[idx].GetSegment(segmentID)
		if err == nil {
			segment.setSurfaceType(surfaceType, probability)
			segment.setLastModified(&timestamp)
			db.roads[idx].setLastModified(&timestamp)
			return nil
		}
	}

	return fmt.Errorf("Unable to update non existing RoadSegment %s", segmentID)
}

type myDB struct {
	roads []Road
}
