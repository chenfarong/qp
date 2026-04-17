package grpc

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"

	"google.golang.org/protobuf/proto"
)

// Marshal 序列化消息
func Marshal(m proto.Message) ([]byte, error) {
	return proto.Marshal(m)
}

// Unmarshal 反序列化消息
func Unmarshal(data []byte, m proto.Message) error {
	return proto.Unmarshal(data, m)
}

// HandlerInfo 处理器信息
type HandlerInfo struct {
	Instance     any
	MethodName   string
	RequestType  func() interface{}
	ResponseType func() interface{}
}

type HandlerFunc struct {
	Instance     any
	MethodName   string
	RequestType  func() interface{}
	ResponseType func() interface{}
}

// Router 消息路由器
type Router struct {
	MessageHandlerInfo sync.Map
}

// NewRouter 创建消息路由器
func NewRouter() *Router {
	return &Router{
		MessageHandlerInfo: sync.Map{},
	}
}

// RegisterHandler 注册消息处理器（按实例和方法名）
func (r *Router) RegisterHandler(messageID int32, instance any, methodName string, requestType func() interface{}, responseType func() interface{}) {
	// 注册消息信息
	r.MessageHandlerInfo.Store(messageID, HandlerFunc{
		Instance:     instance,
		MethodName:   methodName,
		RequestType:  requestType,
		ResponseType: responseType,
	})

	log.Printf("注册消息处理器: messageID=%d, methodName=%s\n", messageID, methodName)
}

// GetMethodTypes 获取方法的【参数类型】和【返回值类型】
func GetMethodTypes(obj interface{}, methodName string) (
	paramTypes []reflect.Type,
	returnTypes []reflect.Type,
	err error,
) {
	// 获取类型信息（注意：用 TypeOf 而不是 ValueOf）
	typ := reflect.TypeOf(obj)

	// 根据方法名获取方法
	method, ok := typ.MethodByName(methodName)
	if !ok {
		return nil, nil, fmt.Errorf("方法 %s 不存在", methodName)
	}

	// 获取方法签名
	funcType := method.Type

	// 获取参数类型：In(0) 是接收者本身，从 In(1) 开始才是真正参数
	paramCount := funcType.NumIn()
	for i := 1; i < paramCount; i++ {
		paramTypes = append(paramTypes, funcType.In(i))
	}

	// 获取返回值类型
	retCount := funcType.NumOut()
	for i := 0; i < retCount; i++ {
		returnTypes = append(returnTypes, funcType.Out(i))
	}

	return paramTypes, returnTypes, nil
}

// HandleMessage 处理消息
func (r *Router) HandleMessage(ctx context.Context, messageID int32, session string, messageContent []byte) (*[]byte, error) {
	// 获取处理器信息
	infoValue, ok := r.MessageHandlerInfo.Load(messageID)
	if !ok {
		return nil, fmt.Errorf("未找到消息处理器: messageID=%d", messageID)
	}

	handlerFunc := infoValue.(HandlerFunc)

	// 使用 RequestType 创建请求实例并进行 protobuf3 反序列化
	req := handlerFunc.RequestType()
	if err := Unmarshal(messageContent, req.(proto.Message)); err != nil {
		return nil, fmt.Errorf("反序列化请求失败: %v", err)
	}

	// 使用反射调用方法
	instanceValue := reflect.ValueOf(handlerFunc.Instance)
	method := instanceValue.MethodByName(handlerFunc.MethodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("未找到方法: %s", handlerFunc.MethodName)
	}

	// 准备参数：上下文 + 请求实例
	args := []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(req),
	}

	// 调用方法
	results := method.Call(args)
	if len(results) != 2 {
		return nil, fmt.Errorf("方法返回值数量错误: %s", handlerFunc.MethodName)
	}

	// 处理返回值：第一个是响应，第二个是error
	resp := results[0].Interface()
	errValue := results[1].Interface()
	if errValue != nil {
		return nil, errValue.(error)
	}

	// 序列化响应
	result, err := Marshal(resp.(proto.Message))
	if err != nil {
		return nil, fmt.Errorf("序列化响应失败: %v", err)
	}

	return &result, nil
}

// GetHandlerInfo 获取处理器信息
func (r *Router) GetHandlerInfo(messageID int32) (HandlerInfo, bool) {
	infoValue, ok := r.MessageHandlerInfo.Load(messageID)
	if !ok {
		return HandlerInfo{}, false
	}

	info, ok := infoValue.(HandlerInfo)
	return info, ok
}
