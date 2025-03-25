# SOCKS5 代理服务器文档

这是一个使用 Go 语言实现的 SOCKS5 代理服务器，支持 **TCP 和 UDP 转发**，并提供了 **用户名/密码认证**、**日志记录** 和 **配置管理** 功能。

------

## 功能特性

### 1. **SOCKS5 协议支持**

- 支持 SOCKS5 协议的核心功能，包括：
  - **CONNECT** 命令：用于 TCP 代理。
  - **UDP ASSOCIATE** 命令：用于 UDP 代理。
- 支持 IPv4、IPv6 和域名地址类型。

### 2. **用户名/密码认证**

- 支持 SOCKS5 用户名/密码认证（`0x02` 方法）。
- 默认用户名：`user`，默认密码：`pass`。
- 可扩展支持更多认证方式。

### 3. **UDP 数据包解析和转发**

- 支持 UDP 数据包的解析和转发。
- 通过 `UDP ASSOCIATE` 命令创建 UDP 中继服务器。
- 支持 IPv4、IPv6 和域名的 UDP 转发。

### 4. **日志记录**

- 所有操作记录到日志文件中（默认文件为 `socks5.log`）。
- 日志内容包括：
  - 客户端连接信息。
  - 认证结果。
  - 请求处理结果。
  - UDP 数据包转发记录。

### 5. **配置管理**

- 通过 `config.yaml` 文件管理配置。
- 支持以下配置项：
  - `listen_addr`：服务器监听地址（默认 `:1080`）。
  - `log_file`：日志文件路径（默认 `socks5.log`）。

------

## 项目结构

```markdown
socks5-proxy/
├── main.go          # 主程序文件
├── config.yaml       # 配置文件
├── README.md        # 项目文档
└── go.mod           # Go 模块文件
```

------

## 快速开始

### 1. 安装依赖

确保已安装 Go 1.20 或更高版本。

```bash
go mod init socks5-proxy
go mod tidy
```

### 2. 配置服务器

编辑 `config.yaml` 文件，修改以下配置项：

```yaml
listen_addr: ":1080"  # 服务器监听地址
log_file: "socks5.log" # 日志文件路径
```

### 3. 运行服务器

```bash
go run main.go
```

服务器将启动并监听 `:1080` 端口。

### 4. 测试代理

#### TCP 代理测试

使用 `curl` 测试 TCP 代理：

```bash
curl --socks5 user:pass@127.0.0.1:1080 http://example.com
```

#### UDP 代理测试

使用支持 SOCKS5 的客户端测试 UDP 代理（如 `dns2socks` 或其他工具）。

------

## 详细说明

### 1. **SOCKS5 握手**

- 客户端发送支持的认证方法列表。
- 服务器选择支持的方法（如用户名/密码认证）。
- 客户端发送用户名和密码进行认证。

### 2. **TCP 代理**

- 客户端发送 `CONNECT` 请求，包含目标地址和端口。
- 服务器连接到目标地址，并在客户端和目标服务器之间转发数据。

### 3. **UDP 代理**

- 客户端发送 `UDP ASSOCIATE` 请求。
- 服务器创建一个 UDP 中继服务器，并返回中继地址给客户端。
- 客户端通过中继服务器发送 UDP 数据包，服务器负责解析和转发。

### 4. **日志记录**

所有操作记录到日志文件中，包括：

- 客户端连接信息。
- 认证结果（成功或失败）。
- 请求处理结果（目标地址和端口）。
- UDP 数据包转发记录（来源和目标地址）。

### 5. **配置管理**

通过 `config.yaml` 文件管理配置，支持以下配置项：

- `listen_addr`：服务器监听地址（默认 `:1080`）。
- `log_file`：日志文件路径（默认 `socks5.log`）。

------

## 扩展建议

1. **支持更多认证方式**：
   - 实现 GSSAPI 认证。
   - 支持无认证模式。
2. **增强 UDP 转发功能**：
   - 支持 UDP 数据包的加密和压缩。
   - 实现 UDP 数据包的缓存和重传。
3. **性能优化**：
   - 使用连接池管理 TCP 连接。
   - 实现 UDP 数据包的批量处理。
4. **安全性增强**：
   - 支持 TLS 加密通信。
   - 实现 IP 白名单和黑名单。

------

## 示例日志

plaintext

```plaintext
2023/10/10 12:00:00 SOCKS5 server is running on :1080
2023/10/10 12:00:05 New connection from 127.0.0.1:12345
2023/10/10 12:00:05 Authentication successful for user: user
2023/10/10 12:00:05 Forwarding TCP request to example.com:80
2023/10/10 12:00:10 New UDP association from 127.0.0.1:12346
2023/10/10 12:00:10 Forwarded UDP packet from 127.0.0.1:12346 to 8.8.8.8:53
```

------

## 许可证

本项目基于 MIT 许可证开源。详细信息请参阅 LICENSE 文件。

------

## 贡献指南

欢迎提交 Issue 或 Pull Request 来改进本项目！