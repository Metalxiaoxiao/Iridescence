# Iridencense

Iridencense 是一个高效稳定的聊天服务器，使用 Go 语言编写，支持 WebSocket 连接和 HTTP 请求。

## 功能

- 用户登录：用户可以通过 HTTP 请求进行登录，登录成功后将返回一个 token，用户可以使用这个 token 进行后续的操作。
- 第三方登录：第三方服务器可以通过设置authorizedToken来认证已登录的用户
- 发送消息：用户可以通过 WebSocket 连接发送消息。
- 添加朋友：用户可以添加其他用户为朋友，添加成功后将可以看到朋友的在线状态，并可以向朋友发送消息。
- 群聊功能：用户可以创建群聊，群聊可以包含多个用户，用户可以向群聊发送消息。

## 安装

提示:安装前请先配置MySQL服务器

1. 克隆这个仓库到你的本地机器上。

```bash
git clone https://github.com/Metalxiaoxiao/Iridencense.git
```

2. 进入项目目录。

```bash
cd iridencense
```

3. 编译项目。

```bash
go build .
```

## 运行

在项目目录下，运行以下命令：

```bash
./iridencense
```

服务器将在默认端口运行，你可以在配置文件中更改端口。

## 配置

你可以在`config.json`文件中配置服务器的设置，包括日志级别、数据库设置、服务端口等。

```json
{
  "logLevel": 0,
  "port": "8080",
  "DataBaseSettings": {
    "address": "localhost:3306",
    "account": "root",
    "password": "Xx20060902zcs"
  },
  "Rotes": {
    "registerRote": "/register",
    "requestRote": "/request",
    "loginRote": "/login",
    "wsRote": "/ws",
    "uploadRote": "/upload",
    "downloadRote": "/download"
  },
  "WebsocketConnBufferSize": 2048,
  "saltLength": 8,
  "tokenLength": 32,
  "authorizedServerTokens": [
    "token1",
    "token2",
    "token3"
  ],
  "TokenExpiryHours": 24,
  "UserSettings": {
    "DefaultAvatar": "http://127.0.0.1",
    "DefaultSettings": {},
    "DefaultPermission": 0,
    "DefaultFriendList": [
      "1",
      "2"
    ],
    "DefaultGroupList": [
      "3",
      "4"
    ],
    "DefaultNote": "暂无签名",
    "DefaultHomePageData": {}
  }
}
```

## 贡献

如果你有任何问题或建议，欢迎提交 issue 或 pull request。

## 许可证

这个项目暂无许可证，后续会添加。
