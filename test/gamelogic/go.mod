module zagame/test/gamelogic

go 1.25.5

require (
	github.com/gorilla/websocket v1.5.3
	zagame/config v0.0.0
	zagame/pb/golang/gamelogic v0.0.0-00010101000000-000000000000
	zagame/proto v0.0.0-00010101000000-000000000000
)

require google.golang.org/protobuf v1.36.11 // indirect

replace zagame/config => ../../config

replace zagame/pb/golang/gamelogic => ../../pb/golang/gamelogic

replace zagame/proto => ../../proto
