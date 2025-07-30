# IM Client Tool

一个用于连接IM服务器的命令行客户端工具。

## 功能特性

- 连接到IM服务器
- 发送和接收消息
- 支持所有服务器命令
- 实时消息显示
- 优雅的退出处理

## 使用方法

### 基本用法

```bash
# 使用默认设置连接本地服务器
./bin/im-client

# 连接到指定服务器
./bin/im-client -ip 192.168.1.100 -port 8888

# 设置用户名（可选）
./bin/im-client -name "MyUsername"
```

### 命令行参数

- `-ip <地址>`: 服务器IP地址（默认: 127.0.0.1）
- `-port <端口>`: 服务器端口（默认: 8888）
- `-name <用户名>`: 用户名（可选）
- `-h` 或 `--help`: 显示帮助信息

### 支持的聊天命令

连接成功后，你可以使用以下命令：

- `who` - 查看在线用户列表
- `name` - 显示你当前的用户名
- `rename <新用户名>` - 修改你的用户名
- `to <用户名> <消息内容>` - 发送私聊消息
- `quit` 或 `exit` - 退出客户端
- 直接输入文本 - 发送广播消息给所有用户

### 示例会话

```
$ ./bin/im-client
Connecting to IM server at 127.0.0.1:8888...
Connected to IM server!
Commands:
  who                    - List online users
  name                   - Show your current username
  rename <newname>       - Change your username
  to <username> <msg>    - Send private message
  quit                   - Exit the client
  Any other text will be broadcast to all users

> who
online users: [127.0.0.1:54321]
> name
Current name: bacon
> rename Alice
Attempting to rename to: Alice
✓ Name updated to: Alice
You've changed name to "Alice"
> Hello everyone!
[127.0.0.1:54321]Alice: Hello everyone!
> to Bob Hi there!
> quit
Disconnected from server. Goodbye!
```

## 技术实现

- 使用Go语言的net包进行TCP连接
- 并发处理用户输入和服务器消息
- 优雅的错误处理和连接管理
- 支持命令行参数解析

## 构建

```bash
# 单独构建客户端
go build -o bin/im-client ./tools/client

# 或使用构建脚本（会同时构建服务器和客户端）
./scripts/build.sh
```
