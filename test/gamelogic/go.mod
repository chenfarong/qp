module zagame/test/gamelogic

go 1.25.5

require (
	github.com/gorilla/websocket v1.5.1
	zagame/pb/golang/gamelogic v0.0.0-00010101000000-000000000000
	zagame/proto v0.0.0-00010101000000-000000000000
)

replace zagame/pb/golang/gamelogic => ../../pb/golang/gamelogic
replace zagame/proto => ../../proto
