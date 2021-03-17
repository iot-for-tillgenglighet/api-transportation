module github.com/iot-for-tillgenglighet/api-transportation

go 1.15

require (
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/google/uuid v1.1.2
	github.com/iot-for-tillgenglighet/messaging-golang v0.0.0-20201230002037-e79e8e927ae9
	github.com/iot-for-tillgenglighet/ngsi-ld-golang v0.0.0-20210316135358-62e2fe839946
	github.com/kr/text v0.2.0 // indirect
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.7.0
	github.com/streadway/amqp v1.0.0
	golang.org/x/net v0.0.0-20201207224615-747e23833adb // indirect
	golang.org/x/sys v0.0.0-20201207223542-d4d67f95c62d // indirect
	golang.org/x/text v0.3.4 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	gorm.io/driver/postgres v1.0.6
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.20.9
)
