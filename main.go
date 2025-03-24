package socks5_proxy

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

const (
	socks5Version          = 0x05
	authMethodNoAuth       = 0x00
	authMethodUserPass     = 0x02
	authMethodNoAcceptable = 0xFF
	cmdConnect             = 0x01
	cmdUDPAssociate        = 0x03
	addrTypeIPv4           = 0x01
	addrTypeDomain         = 0x03
	addrTypeIPv6           = 0x04
)

func handleHandshake(conn net.Conn) error {
	buf := make([]byte, 2)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		return err
	}

	if buf[0] != socks5Version {
		return errors.New("unsupported SOCKS version")
	}

	nMethods := int(buf[1])
	methods := make([]byte, nMethods)
	_, err = io.ReadFull(conn, methods)
	if err != nil {
		return err
	}

	supportUserPass := false
	for _, method := range methods {
		if method == authMethodUserPass {
			supportUserPass = true
			break
		}
	}

	if supportUserPass {
		_, err = conn.Write([]byte{socks5Version, authMethodUserPass})
	} else {
		_, err = conn.Write([]byte{socks5Version, authMethodNoAcceptable})
		return errors.New("no acceptable authentication method")
	}
	return err
}

func handleUserPassAuth(conn net.Conn) error {
	buf := make([]byte, 2)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		return err
	}

	if buf[0] != 0x01 {
		return errors.New("unsupported authentication version")
	}

	usernameLen := int(buf[1])
	username := make([]byte, usernameLen)
	_, err = io.ReadFull(conn, username)
	if err != nil {
		return err
	}

	buf = make([]byte, 1)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return err
	}

	passwordLen := int(buf[0])
	password := make([]byte, passwordLen)
	_, err = io.ReadFull(conn, password)
	if err != nil {
		return err
	}

	// 验证用户名和密码
	if string(username) != "user" || string(password) != "pass" {
		_, err = conn.Write([]byte{0x01, 0xFF}) // 认证失败
		return errors.New("authentication failed")
	}

	_, err = conn.Write([]byte{0x01, 0x00}) // 认证成功
	return err
}

func handleTCPRequest(conn net.Conn) (string, error) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		return "", err
	}

	addrType := buf[3]
	var addr string

	switch addrType {
	case addrTypeIPv4:
		ip := make([]byte, 4)
		_, err = io.ReadFull(conn, ip)
		if err != nil {
			return "", err
		}
		addr = net.IP(ip).String()
	case addrTypeDomain:
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
	case addrTypeIPv6:
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
	_, err = io.ReadFull(conn, portBuf)
	if err != nil {
		return "", err
	}
	port := binary.BigEndian.Uint16(portBuf)

	addr = net.JoinHostPort(addr, strconv.Itoa(int(port)))

	// 返回成功响应
	_, err = conn.Write([]byte{socks5Version, 0x00, 0x00, addrTypeIPv4, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return "", err
	}

	return addr, nil
}

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
		n, addr, err := udpListener.ReadFrom(buf)
		if err != nil {
			log.Println("UDP read error:", err)
			break
		}

		// 这里可以解析 SOCKS5 UDP 数据包并转发
		log.Printf("Received UDP packet from %s: %s\n", addr, string(buf[:n]))
	}

	return "", nil
}

func handleRequest(conn net.Conn) (string, error) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		return "", err
	}

	if buf[0] != socks5Version {
		return "", errors.New("unsupported SOCKS version")
	}

	switch buf[1] {
	case cmdConnect:
		return handleTCPRequest(conn)
	case cmdUDPAssociate:
		return handleUDPAssociate(conn)
	default:
		return "", errors.New("unsupported command")
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	err := handleHandshake(conn)
	if err != nil {
		log.Println("Handshake error:", err)
		return
	}

	err = handleUserPassAuth(conn)
	if err != nil {
		log.Println("Authentication error:", err)
		return
	}

	addr, err := handleRequest(conn)
	if err != nil {
		log.Println("Request error:", err)
		return
	}

	remote, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println("Dial remote error:", err)
		return
	}
	defer remote.Close()

	go io.Copy(remote, conn)
	io.Copy(conn, remote)
}

func main() {
	listener, err := net.Listen("tcp", ":1080")
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	defer listener.Close()

	fmt.Println("SOCKS5 server is running on :1080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go handleConnection(conn)
	}
}
