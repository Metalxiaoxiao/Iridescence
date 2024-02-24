## WebSocket API 文档

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

# Iridencense HTTP API 文档

以下是 Iridencense HTTP API 支持的请求和响应：

## 注册用户

- **URL**: `/register`
- **Method**: `POST`
- **Content-Type**: `application/json` 或 `application/x-www-form-urlencoded`
- **Request Body** (JSON):
  - `userName`: 用户名，不超过10个字符
  - `password`: 密码，长度在8到100个字符之间，必须包含大小写字母和数字
- **Response**: 用户唯一的自增ID

## 用户登录

- **URL**: `/login`
- **Method**: `POST`
- **Content-Type**: `application/x-www-form-urlencoded`
- **Request Body** (JSON):
  - `userId`: 用户ID
  - `password`: 用户密码
- **Response**: 用户的 token

## 获取用户信息

- **URL**: `/request`
- **Method**: `POST`
- **Content-Type**: `application/json` 或 `application/x-www-form-urlencoded`
- **Request Body** (JSON):
  - `token`: 用户的 token
  - `command`: 指令，可以是`getUserData`或`getUserDataByID`
  - `target` (可选): 目标用户的ID，仅在`command`为`getUserDataByID`时使用
- **Response**: 用户的信息，包括用户名、头像、备注、权限和好友列表

## 验证用户 token

- **URL**: `/request`
- **Method**: `POST`
- **Content-Type**: `application/json` 或 `application/x-www-form-urlencoded`
- **Request Body** (JSON):
  - `token`: 服务器的 token
  - `command`: 指令，必须为`verifyToken`
  - `targetToken`: 要验证的用户 token
- **Response**: 用户的信息，包括用户名、头像、备注、权限和好友列表

## 发布帖子

- **URL**: `/request`
- **Method**: `POST`
- **Content-Type**: `application/json`
- **Request Body** (JSON):
  - `token`: 用户的 token
  - `command`: 指令，必须为`publishPost`
  - `userID`: 用户ID
  - `content`: 帖子内容
- **Response**: 发布成功的确认

## 获取某个帖子

- **URL**: `/request`
- **Method**: `POST`
- **Content-Type**: `application/json`
- **Request Body** (JSON):
  - `token`: 用户的 token
  - `command`: 指令，必须为`getPost`
  - `postID`: 帖子ID
- **Response**: 指定的帖子

## 获取用户的帖子

- **URL**: `/request`
- **Method**: `POST`
- **Content-Type**: `application/json`
- **Request Body** (JSON):
  - `token`: 用户的 token
  - `command`: 指令，必须为`getUserPosts`
  - `userID`: 用户ID
  - `startTime`: 开始时间 (Unix timestamp)
  - `endTime`: 结束时间 (Unix timestamp)
- **Response**: 用户在指定时间段内的所有帖子的列表

## 获取所有帖子

- **URL**: `/request`
- **Method**: `POST`
- **Content-Type**: `application/json`
- **Request Body** (JSON):
  - `token`: 用户的 token
  - `command`: 指令，必须为`getPosts`
  - `startTime`: 开始时间 (Unix timestamp)
  - `endTime`: 结束时间 (Unix timestamp)
- **Response**: 指定时间段内的所有帖子的列表

## 错误响应

- **400 Bad Request**: 请求的参数无效或缺失
- **401 Unauthorized**: 提供的 token 无效
- **500 Internal Server Error**: 服务器内部错误

## 注意

- 所有的请求都必须是 POST 请求
- 所有的参数都应该在请求体中以 JSON 或表单的形式提供
- 所有的响应都是纯文本，除了获取用户信息的响应是 JSON 格式
- 用户的 token 在登录成功后由服务器生成并返回，之后的所有请求都应该在请求头中提供这个 token
- 服务器的 token 应该在配置文件中提供，并且只能用于验证用户 token 的请求

请注意，API暂不完善，仍在开发中