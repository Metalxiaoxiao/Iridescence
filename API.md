## API 文档

以下是 Iridencense 支持的 WebSocket 命令和对应的 JSON 数据格式：

### 登录 - `login`

请求：

```json
{
  "command": "login",
  "userId": 1,
  "password": "password123"
}
```

响应：

```json
{
  "state": true,
  "message": "登录成功"
}
```

### 注册 - `signUp`

请求：

```json
{
  "command": "signUp",
  "userid": 1,
  "password": "password123"
}
```

### 发送消息 - `sendMessage`

请求：

```json
{
  "command": "sendMessage",
  "targetId": 2,
  "requestId": 1,
  "messageBody": "Hello, world!",
  "time": 1631846000
}
```

响应：

```json
{
  "requestId": 1,
  "messageId": 1,
  "time": 1631846000,
  "state": 1
}
```

### 添加好友 - `addFriend`

请求：

```json
{
  "command": "addFriend",
  "friendId": 2
}
```

### 删除好友 - `deleteFriend`

请求：

```json
{
  "command": "deleteFriend",
  "friendId": 2
}
```

### 获取用户数据 - `getUserData`

请求：

```json
{
  "command": "getUserData",
  "userId": 1
}
```

响应：

```json
{
  "userId": 1,
  "userName": "张三",
  "userAvatar": "http://example.com/avatar.jpg",
  "userNote": "这是一个备注",
  "userPermission": 1,
  "userFriendList": "[2,3,4]"
}
```

### 更改头像 - `changeAvatar`

请求：

```json
{
  "command": "changeAvatar",
  "newAvatar": "http://example.com/new_avatar.jpg"
}
```

响应：

```json
{
  "userId": 1,
  "newAvatar": "http://example.com/new_avatar.jpg",
  "success": true
}
```

### 获取与用户的消息 - `getMessagesWithUser`

请求：

```json
{
  "command": "getMessagesWithUser",
  "otherUserId": 2,
  "startTime": 1631846000,
  "endTime": 1631932400
}
```

响应：

```json
{
  "userId": 1,
  "messages": [
    {
      "messageId": 1,
      "senderId": 1,
      "receiverId": 2,
      "time": 1631846000,
      "messageBody": "Hello, world!",
      "messageType": 1
    }
  ]
}
```

### 创建群聊 - `createGroup`

请求：

```json
{
  "command": "createGroup",
  "groupName": "新群聊",
  "groupExplaination": "这是一个新的群聊"
}
```

响应：

```json
{
  "groupId": 1,
  "success": true
}
```

### 解散群聊 - `breakGroup`

请求：

```json
{
  "command": "breakGroup",
  "groupId": 1
}
```

响应：

```json
{
  "groupId": 1,
  "success": true
}
```

### 发送群消息 - `sendGroupMessage`

请求：

```json
{
  "command": "sendGroupMessage",
  "groupId": 1,
  "messageBody": "Hello, group!",
  "requestId": 1
}
```

响应：

```json
{
  "requestId": 1,
  "messageId": 1,
  "time": 1631846000,
  "state": 1
}
```

API暂不完善，仍在开发中