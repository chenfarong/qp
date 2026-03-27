package proto

import (
	"github.com/aoyo/qp/proto"
)

// Serialize 序列化消息
func Serialize(msg *proto.Message) ([]byte, error) {
	return msg.Marshal()
}

// Deserialize 反序列化消息
func Deserialize(data []byte) (*proto.Message, error) {
	msg := &proto.Message{}
	err := msg.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
