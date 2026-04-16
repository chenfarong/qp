module zagame/inside/gamelogic

go 1.25.5

require (
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.33.0
	zagame/pb/golang/gamelogic v0.0.0-00010101000000-000000000000
)

replace zagame/pb/golang/gamelogic => ../../pb/golang/gamelogic

require (
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
)
