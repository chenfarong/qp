const WebSocket = require('ws');

// 创建WebSocket连接
const ws = new WebSocket('ws://localhost:8080/ws');

// 连接打开时
ws.on('open', function open() {
  console.log('WebSocket connection opened');
  
  // 发送文本消息
  ws.send(JSON.stringify({
    type: 'test',
    message: 'Hello from client'
  }));
});

// 收到消息时
ws.on('message', function incoming(data) {
  console.log('Received:', data);
});

// 连接关闭时
ws.on('close', function close() {
  console.log('WebSocket connection closed');
});

// 连接错误时
ws.on('error', function error(err) {
  console.error('WebSocket error:', err);
});
