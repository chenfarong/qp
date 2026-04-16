module zgame/internet/gateway

go 1.25.5

require (
	github.com/gorilla/websocket v1.5.3
	github.com/lib/pq v1.10.9
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.33.0
	zgame/config v0.0.0-00010101000000-000000000000
	zgame/database v0.0.0-00010101000000-000000000000
)

replace zgame/config => ../../config
replace zgame/database => ../../database

require (
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
