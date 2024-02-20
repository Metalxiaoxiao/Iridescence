# Iridencense

Iridencense 是一个高效稳定的聊天服务器，使用 Go 语言编写，支持 WebSocket 连接和 HTTP 请求。

## 功能

- 用户登录：用户可以通过 HTTP 请求进行登录，登录成功后将返回一个 token，用户可以使用这个 token 进行后续的操作。
- 发送消息：用户可以通过 WebSocket 连接发送消息。
- 添加朋友：用户可以添加其他用户为朋友，添加成功后将可以看到朋友的在线状态，并可以向朋友发送消息。
- 群聊功能：用户可以创建群聊，群聊可以包含多个用户，用户可以向群聊发送消息。
- 更改设置：用户可以更改自己的设置，包括用户名、头像、权限等。

## 安装

1. 克隆这个仓库到你的本地机器上。

```bash
git clone https://github.com/Metalxiaoxiao/iridencense.git
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
    "LogLevel": "info",
    "DataBaseSettings": {
        "Account": "default_account",
        "Password": "default_password",
        "Address": "localhost:3306"
    },
    "Port": "8080",
    "WebSocketServiceRote": "/ws",
    "RegisterServiceRote": "/register",
    "LoginServiceRote": "/login",
    "UploadServiceRote": "/upload",
    "DownloadServiceRote": "/download"
}
```

## 贡献

如果你有任何问题或建议，欢迎提交 issue 或 pull request。

## 许可证

这个项目暂无许可证，后续会添加。
