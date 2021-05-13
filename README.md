# Introduction 

This service is responsible for storing road and road segment information and provide it to consumers via an API.

## Deprecation Notice

This repository is deprecated and the code has moved to https://github.com/diwise/api-transportation

# Building, tagging and running with Docker

`docker build -f deployments/Dockerfile -t iot-for-tillgenglighet/api-transportation:latest .`
`docker run -it -p 8880:8880 iot-for-tillgenglighet/api-transportation:latest`

# Request data from the service

Get all roadsegments within a rectangle described by three GeoJSON positions in [lon,lat]-format:

`http://localhost:8484/ngsi-ld/v1/entities?type=RoadSegment&georel=within&geometry=Polygon&coordinates=[[17.230700,62.430242],[17.444075,62.353557],[17.444075,62.353557]]`

Get all roadsegments within a distance (30 meters) from a [lon,lat] point:

`http://localhost:8484/ngsi-ld/v1/entities?type=RoadSegment&georel=near;maxDistance==30&geometry=Point&coordinates=[17.342553,62.377022]`
