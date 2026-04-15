package protocol

import (
	"bytes"
	"encoding/binary"
	"io"
)

// MessageType 消息类型
type MessageType byte

const (
	MessageTypeRequest  MessageType = 'R' // 请求消息
	MessageTypeResponse MessageType = 'S' // 响应消息
	MessageTypeNotify   MessageType = 'N' // 通知消息
)

// CompressFlag 压缩标志
type CompressFlag byte

const (
	CompressFlagNone   CompressFlag = 'N' // 不压缩
	CompressFlagGzip   CompressFlag = 'G' // Gzip压缩
	CompressFlagSnappy CompressFlag = 'S' // Snappy压缩
)

// Packet 数据包结构
type Packet struct {
	MessageType  MessageType  // 消息类型
	CompressFlag CompressFlag // 压缩标志
	MessageID    int32        // 消息号
	Data         []byte       // 消息数据
}

// Encode 编码数据包
func Encode(msgType MessageType, compressFlag CompressFlag, msgID int32, data []byte) ([]byte, error) {
	// 创建缓冲区
	buf := new(bytes.Buffer)

	// 写入消息类型
	if err := binary.Write(buf, binary.LittleEndian, msgType); err != nil {
		return nil, err
	}

	// 写入压缩标志
	if err := binary.Write(buf, binary.LittleEndian, compressFlag); err != nil {
		return nil, err
	}

	// 写入消息号
	if err := binary.Write(buf, binary.LittleEndian, msgID); err != nil {
		return nil, err
	}

	// 写入消息数据
	if err := binary.Write(buf, binary.LittleEndian, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decode 解码数据包
func Decode(reader io.Reader) (*Packet, error) {
	// 读取消息类型
	var msgType MessageType
	if err := binary.Read(reader, binary.LittleEndian, &msgType); err != nil {
		return nil, err
	}

	// 读取压缩标志
	var compressFlag CompressFlag
	if err := binary.Read(reader, binary.LittleEndian, &compressFlag); err != nil {
		return nil, err
	}

	// 读取消息号
	var msgID int32
	if err := binary.Read(reader, binary.LittleEndian, &msgID); err != nil {
		return nil, err
	}

	// 读取消息数据（直到EOF）
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return &Packet{
		MessageType:  msgType,
		CompressFlag: compressFlag,
		MessageID:    msgID,
		Data:         data,
	}, nil
}
