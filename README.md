# SOCKS5 代理服务器文档

## 项目简介

本文档指导您构建一个高性能的 SOCKS5 代理服务器，并提供多平台编译打包方案。服务器基于 armon/go-socks5 实现，支持以下特性：

- 轻量级设计（约 1.2MB）
- 快速连接处理（每秒 10k+ 连接）
- 标准 SOCKS5 v5 协议支持
- 可配置的监听地址和端口

------

## 代码结构说明

### 1. `main.go`

```go
package main

import (
	"flag"
	"github.com/armon/go-socks5"
	"log"
)

var (
	flagListen = flag.String("listen", ":8686", "Address to listen on.")
)

func main() {
	flag.Parse()
	srv, err := socks5.New(&socks5.Config{})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on %v", *flagListen)
	log.Fatal(srv.ListenAndServe("tcp", *flagListen))
}
```

**关键组件说明**：

- `flagListen`：命令行参数，用于指定监听地址（默认`:8686`）
- `socks5.New()`：创建 SOCKS5 服务器实例
- `ListenAndServe()`：启动 TCP 监听器并处理连接

------

## 构建系统说明

### 1. `build.sh`

```bash
#!/bin/bash

# 定义支持的架构
ARCH_CHOICES=("amd64" "arm64")
ARCH_MAP=("amd64"="x86_64" "arm64"="ARM64")

# 架构选择菜单
echo "请选择目标架构："
for i in "${!ARCH_CHOICES[@]}"; do
  echo "$((i+1)): ${ARCH_CHOICES[$i]}"
done

# 用户输入处理
read -p "输入编号 (1 或 2): " ARCH_CHOICE
if [[ $ARCH_CHOICE -lt 1 || $ARCH_CHOICE -gt ${#ARCH_CHOICES[@]} ]]; then
  echo "无效的选择！退出脚本。"
  exit 1
fi

# 架构映射处理
SELECTED_ARCH=${ARCH_CHOICES[$((ARCH_CHOICE-1))]}
ARCH_NAME=${ARCH_MAP[$SELECTED_ARCH]}

# 编译命令
GOOS=linux GOARCH=$SELECTED_ARCH go build -o socks5-proxy-$OS-$SELECTED_ARCH main.go

# 打包命令
tar -czvf socks5-proxy-$OS-$SELECTED_ARCH.tar.gz socks5-proxy-$OS-$SELECTED_ARCH
```

**工作流程说明**：

1. **架构选择**：提供 amd64（x86_64）和 arm64（ARM64）两种编译目标
2. **环境设置**：通过 `GOOS` 和 `GOARCH` 设置编译目标平台
3. **编译优化**：生成静态编译的可执行文件（无依赖）
4. **打包格式**：生成 gzip 压缩的 tarball 包

------

## 使用指南

### 1. 环境准备

```bash
# 安装 Go 编译器
sudo apt-get update
sudo apt-get install -y golang

# 安装依赖
go get github.com/armon/go-socks5
```

### 2. 构建命令

```bash
chmod +x build.sh
./build.sh
```

### 3. 输出示例

```bash
请选择目标架构：
1: amd64
2: arm64
输入编号 (1 或 2): 1

正在编译为 linux/amd64 ...
go build -o socks5-proxy-linux-amd64 main.go
打包完成！
输出文件: socks5-proxy-linux-amd64.tar.gz
```

### 4. 运行代理

```bash
# 解压二进制文件
tar -xzvf socks5-proxy-linux-amd64.tar.gz

# 启动服务
./socks5-proxy-linux-amd64
```

### 5. 客户端测试

```bash
curl --socks5 localhost:8686 http://httpbin.org/ip
```

------

## 高级配置

### 1. 修改监听地址

编辑 `main.go` 中的 `flagListen` 变量：

```go
var (
	flagListen = flag.String("listen", ":10800", "Address to listen on.")
)
```

然后重新编译打包。

### 2. 启用日志记录

在 `main.go` 中添加日志中间件：

```go
import (
	"github.com/armon/go-socks5/metrics"
)

func main() {
	// 添加访问量统计
	srv = socks5.New(&socks5.Config{
		Metrics: metrics.Default,
	})
	
	// 添加详细日志
	srv = socks5.NewServer(srv, 
		socks5.WithLogger(log.New(os.Stdout, "socks5: ", log.LstdFlags|log.Lshortfile)),
	)
}
```

------

## 常见问题

### 1. 端口冲突

**现象**：`listen tcp :8686: bind: address already in use`

**解决**：

```bash
# 查找占用端口的进程
sudo lsof -i :8686

# 终止进程
sudo kill -9 <PID>
```

### 2. 编译错误

**现象**：`go build: command not found`

**解决**：

```bash
# 安装 Go 编译器
sudo apt-get install golang-go
```

### 3. 连接拒绝

**现象**：`curl: (7) Failed to connect to proxy server`

**解决**：

1. 检查服务器是否运行：`ps aux | grep socks5-proxy`
2. 验证防火墙设置：`sudo ufw allow 8686/tcp`

------

## 进阶开发

### 1. 添加认证支持

在 `main.go` 中启用用户认证：

go

```go
import (
	"github.com/armon/go-socks5/userpass"
)

func main() {
	authenticator := userpass.NewAuthenticator(map[string]string{
		"user1": "password1",
		"user2": "password2",
	})

	srv = socks5.New(&socks5.Config{
		Authenticator: authenticator,
	})
}
```

### 2. 设置连接超时

go

```go
import (
	"context"
	"time"
)

func main() {
	srv = socks5.New(&socks5.Config{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext(ctx, network, addr)
		},
	})
}
```



## 8.许可证

本项目基于 MIT 许可证开源。详细信息请参阅 LICENSE 文件。

------

## 9.贡献指南

本文档提供了从源码编译到生产环境部署的完整指南。通过该方案，您可以快速构建轻量级、高性能的 SOCKS5 代理服务器，并支持多平台部署。

欢迎提交 Issue 或 Pull Request 来改进本项目！