package main

import (
	"encoding/binary"  // 用于处理二进制数据的编码和解码
	"errors"           // 用于定义错误
	"gopkg.in/yaml.v3" // 用于解析 YAML 配置文件
	"io"               // 提供 I/O 操作接口
	"log"              // 用于日志记录
	"net"              // 提供网络操作接口
	"os"               // 提供操作系统相关功能
	"strconv"          // 用于字符串和数字的转换
)

// 定义常量
const (
	socks5Version          = 0x05 // SOCKS5 协议版本号
	authMethodNoAuth       = 0x00 // 无认证方法
	authMethodUserPass     = 0x02 // 用户名/密码认证方法
	authMethodNoAcceptable = 0xFF // 无支持的认证方法
	cmdConnect             = 0x01 // CONNECT 命令（用于 TCP 代理）
	cmdUDPAssociate        = 0x03 // UDP ASSOCIATE 命令（用于 UDP 代理）
	addrTypeIPv4           = 0x01 // IPv4 地址类型
	addrTypeDomain         = 0x03 // 域名地址类型
	addrTypeIPv6           = 0x04 // IPv6 地址类型
)

// Config 定义配置结构体
type Config struct {
	ListenAddr string `yaml:"listen_addr"` // 服务器监听地址
	LogFile    string `yaml:"log_file"`    // 日志文件路径
	Username   string `yaml:"username"`    // 用户名
	Password   string `yaml:"password"`    // 密码
}

var cfg Config // 全局配置变量

// init 初始化函数，读取配置文件并设置日志
func init() {
	// 读取配置文件
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	// 解析 YAML 配置文件
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	// 初始化日志文件
	logFile, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(logFile) // 设置日志输出到文件
}

// handleHandshake 处理 SOCKS5 握手
func handleHandshake(conn net.Conn) error {
	buf := make([]byte, 2)
	_, err := io.ReadFull(conn, buf) // 读取客户端发送的握手请求
	if err != nil {
		return err
	}

	if buf[0] != socks5Version { // 检查协议版本
		return errors.New("unsupported SOCKS version")
	}

	nMethods := int(buf[1]) // 获取支持的认证方法数量
	methods := make([]byte, nMethods)
	_, err = io.ReadFull(conn, methods) // 读取支持的认证方法
	if err != nil {
		return err
	}

	// 打印支持的认证方法
	log.Printf("Supported methods: %v", methods)

	supportNoAuth := false
	supportUserPass := false

	for _, method := range methods { // 检查支持的认证方法
		if method == authMethodNoAuth {
			supportNoAuth = true
		}
		if method == authMethodUserPass {
			supportUserPass = true
		}
	}

	if supportUserPass {
		_, err = conn.Write([]byte{socks5Version, authMethodUserPass}) // 返回支持用户名/密码认证
	} else if supportNoAuth {
		_, err = conn.Write([]byte{socks5Version, authMethodNoAuth}) // 返回支持无认证
	} else {
		_, err = conn.Write([]byte{socks5Version, authMethodNoAcceptable}) // 返回无支持的认证方法
		return errors.New("no acceptable authentication method")
	}
	return err
}

// handleUserPassAuth 处理用户名/密码认证
func handleUserPassAuth(conn net.Conn) error {
	buf := make([]byte, 2)
	_, err := io.ReadFull(conn, buf) // 读取认证版本和用户名长度
	if err != nil {
		return err
	}

	if buf[0] != 0x01 { // 检查认证版本
		return errors.New("unsupported authentication version")
	}

	usernameLen := int(buf[1]) // 获取用户名长度
	usernameBuf := make([]byte, usernameLen)
	_, err = io.ReadFull(conn, usernameBuf) // 读取用户名
	if err != nil {
		return err
	}

	buf = make([]byte, 1)
	_, err = io.ReadFull(conn, buf) // 读取密码长度
	if err != nil {
		return err
	}

	passwordLen := int(buf[0]) // 获取密码长度
	passwordBuf := make([]byte, passwordLen)
	_, err = io.ReadFull(conn, passwordBuf) // 读取密码
	if err != nil {
		return err
	}

	// 验证用户名和密码
	if string(usernameBuf) != cfg.Username || string(passwordBuf) != cfg.Password {
		_, err = conn.Write([]byte{0x01, 0xFF}) // 认证失败
		return errors.New("authentication failed")
	}

	_, err = conn.Write([]byte{0x01, 0x00}) // 认证成功
	return err
}

// handleTCPRequest 处理 TCP 请求
func handleTCPRequest(conn net.Conn) (string, error) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(conn, buf) // 读取请求头
	if err != nil {
		return "", err
	}

	addrType := buf[3] // 获取地址类型
	var addr string

	switch addrType {
	case addrTypeIPv4: // IPv4 地址
		ip := make([]byte, 4)
		_, err = io.ReadFull(conn, ip)
		if err != nil {
			return "", err
		}
		addr = net.IP(ip).String()
	case addrTypeDomain: // 域名地址
		domainLenBuf := make([]byte, 1)
		_, err = io.ReadFull(conn, domainLenBuf)
		if err != nil {
			return "", err
		}
		domainLen := int(domainLenBuf[0])
		domain := make([]byte, domainLen)
		_, err = io.ReadFull(conn, domain)
		if err != nil {
			return "", err
		}
		addr = string(domain)
	case addrTypeIPv6: // IPv6 地址
		ip := make([]byte, 16)
		_, err = io.ReadFull(conn, ip)
		if err != nil {
			return "", err
		}
		addr = net.IP(ip).String()
	default:
		return "", errors.New("unsupported address type")
	}

	portBuf := make([]byte, 2)
	_, err = io.ReadFull(conn, portBuf) // 读取端口号
	if err != nil {
		return "", err
	}
	port := binary.BigEndian.Uint16(portBuf)

	addr = net.JoinHostPort(addr, strconv.Itoa(int(port))) // 拼接地址和端口

	// 返回成功响应
	_, err = conn.Write([]byte{socks5Version, 0x00, 0x00, addrTypeIPv4, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return "", err
	}

	return addr, nil
}

// handleUDPAssociate 处理 UDP 请求
func handleUDPAssociate(conn net.Conn) (string, error) {
	// 创建一个 UDP 监听器
	udpAddr := "0.0.0.0:0"
	udpListener, err := net.ListenPacket("udp", udpAddr)
	if err != nil {
		return "", err
	}
	defer udpListener.Close()

	// 返回 UDP 中继地址给客户端
	_, port, err := net.SplitHostPort(udpListener.LocalAddr().String())
	if err != nil {
		return "", err
	}
	portInt, _ := strconv.Atoi(port)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(portInt))

	_, err = conn.Write([]byte{socks5Version, 0x00, 0x00, addrTypeIPv4, 0, 0, 0, 0, portBytes[0], portBytes[1]})
	if err != nil {
		return "", err
	}

	// 处理 UDP 数据包
	buf := make([]byte, 65536)
	for {
		n, addr, err := udpListener.ReadFrom(buf) // 读取 UDP 数据包
		if err != nil {
			log.Println("UDP read error:", err)
			break
		}

		// 解析 SOCKS5 UDP 数据包
		if n < 10 {
			log.Println("Invalid UDP packet")
			continue
		}

		// 解析目标地址和端口
		destAddrType := buf[3]
		var destAddr string
		var destPort uint16

		switch destAddrType {
		case addrTypeIPv4:
			destAddr = net.IP(buf[4:8]).String()
			destPort = binary.BigEndian.Uint16(buf[8:10])
		case addrTypeDomain:
			domainLen := int(buf[4])
			destAddr = string(buf[5 : 5+domainLen])
			destPort = binary.BigEndian.Uint16(buf[5+domainLen : 7+domainLen])
		case addrTypeIPv6:
			destAddr = net.IP(buf[4:20]).String()
			destPort = binary.BigEndian.Uint16(buf[20:22])
		default:
			log.Println("Unsupported address type in UDP packet")
			continue
		}

		dest := net.JoinHostPort(destAddr, strconv.Itoa(int(destPort))) // 拼接目标地址和端口

		// 打印 UDP 请求信息
		log.Printf("UDP request from %s to %s\n", addr, dest)

		// 转发 UDP 数据包
		udpConn, err := net.Dial("udp", dest)
		if err != nil {
			log.Printf("Failed to dial UDP destination %s: %v\n", dest, err)
			continue
		}
		defer udpConn.Close()

		_, err = udpConn.Write(buf[10:n]) // 转发数据包
		if err != nil {
			log.Printf("Failed to forward UDP packet to %s: %v\n", dest, err)
			continue
		}
	}

	return "", nil
}

// handleRequest 处理客户端请求
func handleRequest(conn net.Conn) (string, error) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(conn, buf) // 读取请求头
	if err != nil {
		return "", err
	}

	if buf[0] != socks5Version { // 检查协议版本
		return "", errors.New("unsupported SOCKS version")
	}

	switch buf[1] {
	case cmdConnect: // 处理 TCP 请求
		return handleTCPRequest(conn)
	case cmdUDPAssociate: // 处理 UDP 请求
		return handleUDPAssociate(conn)
	default:
		return "", errors.New("unsupported command")
	}
}

// handleConnection 处理客户端连接
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// 打印客户端连接信息
	clientAddr := conn.RemoteAddr().String()
	log.Printf("New connection from: %s\n", clientAddr)

	// 处理握手
	err := handleHandshake(conn)
	if err != nil {
		log.Println("Handshake error:", err)
		return
	}

	// 处理认证
	err = handleUserPassAuth(conn)
	if err != nil {
		log.Println("Authentication error:", err)
		return
	}

	// 处理请求
	addr, err := handleRequest(conn)
	if err != nil {
		log.Println("Request error:", err)
		return
	}

	// 打印目标地址
	log.Printf("Forwarding request to: %s\n", addr)

	// 连接到目标服务器
	remote, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println("Dial remote error:", err)
		return
	}
	defer remote.Close()

	// 转发数据
	go io.Copy(remote, conn)
	io.Copy(conn, remote)
}

// main 主函数，启动 SOCKS5 服务器
func main() {
	// 监听 TCP 端口
	listener, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}
	defer listener.Close()

	log.Printf("SOCKS5 server is running on %s\n", cfg.ListenAddr)

	// 接受客户端连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go handleConnection(conn) // 处理连接
	}
}
