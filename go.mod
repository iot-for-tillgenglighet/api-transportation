module github.com/iot-for-tillgenglighet/api-transportation

go 1.15

require (
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/iot-for-tillgenglighet/ngsi-ld-golang v0.0.0-20201113145248-1684fc0ab74c
	github.com/kr/pretty v0.1.0 // indirect
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1 // indirect
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7 // indirect
	golang.org/x/sys v0.0.0-20200223170610-d5e6a3e2c0ae // indirect
	golang.org/x/text v0.3.3 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gorm.io/gorm v1.20.6
)

replace github.com/99designs/gqlgen => github.com/marwan-at-work/gqlgen v0.0.0-20200107060600-48dc29c19314
