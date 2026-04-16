一个游戏服务器架构
支持百万级同时在线玩家
golang语言实现的
版本更新，支持灰度，资源配表热更，功能单独更
数据库采用 postgresql
配置管理采用 etcd
所有服务器集中配置 config.yml 一个文件中


## 服务器职责清单

ssoauth
负责账号登录，账号注册等管理
支持jwt标准
登录成功后，给客户端发送一个session字符串32个小写字母
对外提供一个标准的http协议服务器
session有效期管理


gateway
对外提供websocket长连接请求
对内提供一个grpc服务器，供其他服务器转发消息用
其他服务器向gateway注册消息处理整数段，gateway收到消息后，根据消息号，找到对应的服务器，将消息通过grpc连接转发给对应的服务器
客户端websocket连接成功后，向此服务器报告自己的session值
后面其他服务器发送给这个session的消息都将由这个gateway服发送给客户端
一组游戏服，配置一个gateway服，其他逻辑处理服之间的通信都通过此服
将收到的请求通过grpc转发给其他服务器
和客户端之间的通信编码为消息类型（char），标志（char）压缩等，消息号（int32），消息内容
当客户端获得session后，再转发给其他服务器的时候，在通信内容最前面加上这个session字符串
消息协议内部以protobuf3为标准


## gamelogic
负责处理玩家角色的游戏逻辑
有一个grpc的客户端，连接到gateway服务器
还有一个http的监听端口，接受http来的post协议请求，post中的内容格式为json方式
一个角色数据在数据库中就一个文档，并且有一个唯一的角色编号（aid）32个小写字母
代码目录用功能模块来区分
消息处理路由通过消息号来进行
程序发生painc的时候，不要崩溃，以崩溃类型的打印日志
日志通过udp方式发送给本网段中的监听端口
内存中保存当前在线的 actor_id和session的map
在线角色数据在内存中，直到收到gateway对此角色下线通知后，才从内存中删除
一个角色绑定一个微线程，一个channel
handler中的每个消息处理方法定义是 上下文，protobuf3的请求，返回protobuf3的返回请求和错误





### gamelogic目录结构
inside\gamelogic\
	base
		handler.go
		model.go
		service.go
	bag
		handler.go
		model.go
		service.go
	hero
		handler.go
		model.go
		service.go
	main.go

目录根据 proto\gamelogic 下的protobuf3定义文件来，一个文件一个目录，一个目录下有 handler.go,service.go,model.go 三个文件;
handler.go 中根据协议定义文件中的 Request 结尾的消息都有一个对应的方法，方法参数是 处理上下文，和生成的pb结构，方法的返回值就是离这个请求最近的下方一个Response生成的pb结构指针和error两个值


### 每个service都要实现如下接口

type ServiceMessageResult struct {
	serviceName    string `json:"serviceName"`
	messageContent []byte `json:"messageContent"`
}

type Service interface {
	OnStartup()
	OnShutdown()
	//内部消息处理接口
	HandleMessage(messageID string, actorID string, messageContent []byte, results []ServiceMessageResult)
}


### 每个消息处理都要实现如下接口

type Handler interface {
	RegisterHandlers()
	OnActorUse(actorID string)
	OnActorLogout(actorID string)
	OnActorOnline(actorID string)
	OnActorOffline(actorID string)
}

在 main.go 中统一注册调用这些接口


loger
负责接收此组服务器的警告，错误，崩溃三类日志
启动一个udp的监听
启动一个tcp的监听，tcp连接成功后，一个客户端连接一个线程，将收到的所有日志都发送给tcp连接客户端
单独一个线程将收到的日志写到文件中去



代码目录结构
bin 放所有编译好的执行文件
vendor 所有第三方的依赖都下载到这个目录下
kits 项目中用到的工具程序
test 功能测试代码
res 项目中用到的资源配置文件
proto 和客户端之间的protobuf3定义文件
protosvr 服务器之间用的protobuf3定义文件
outside 对外的服务器代码，gateway，ssoauth
inside 内部的服务器代码，gamelogic，loger
xlsx 项目中用到的配置文件，在转换后就放到res目录下
pb 根据消息定义文件生成的不同程序语言用的序列化/反序列化目录，例如：pb\golang\gamelogic\bag.pb.go 这样存放


proto 目录下根据不同的服务器文件夹，放置那个服务器的消息定义文件
