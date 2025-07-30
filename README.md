# IM System

一个基于Go语言实现的简单即时通讯系统。

## 功能特性

- 多用户在线聊天
- 用户重命名
- 私聊功能
- 广播消息
- 在线用户查询
- 自动超时踢出

## 项目结构

```
IM-System/
├── cmd/
│   └── main.go              # 程序入口
├── internal/
│   ├── server.go            # 服务器核心逻辑
│   ├── user.go              # 用户管理
│   └── message.go           # 消息处理
├── scripts/
│   └── build.sh             # 构建脚本
├── bin/                     # 编译输出目录
├── go.mod                   # Go模块文件
├── go.sum                   # 依赖版本锁定
└── README.md                # 项目说明
```

## 快速开始

### 构建项目

```bash
# 使用构建脚本
./scripts/build.sh

# 或者直接使用go build
go build -o bin/im-server ./cmd
```

### 运行服务器

```bash
./bin/im-server
```

服务器将在 `127.0.0.1:8888` 启动。

### 连接服务器

使用telnet或其他TCP客户端连接：

```bash
telnet 127.0.0.1 8888
```

## 支持的命令

- `who` - 查看在线用户列表
- `rename <新用户名>` - 修改用户名
- `to <用户名> <消息内容>` - 发送私聊消息
- 直接输入消息 - 发送广播消息

## 技术栈

- Go 1.24.4
- go.uber.org/zap (日志库)

## 开发

### 依赖管理

```bash
# 下载依赖
go mod download

# 整理依赖
go mod tidy
```

### 代码结构说明

- `cmd/main.go`: 程序入口点
- `internal/server.go`: 服务器核心逻辑，处理连接和消息分发
- `internal/user.go`: 用户相关操作，如上线、下线、发送消息
- `internal/message.go`: 消息处理器，实现各种命令的处理逻辑

## 许可证

MIT License
