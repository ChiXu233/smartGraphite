package service

import (
	"fmt"
	"log"
	"net"
)

func SocketServer() {
	listen, err := net.Listen("tcp", "0.0.0.0:8002") //代表监听的地址端口
	if err != nil {
		fmt.Println("listen failed, err:", err)
		return
	}
	fmt.Println("正在等待建立连接.....", listen.Addr())
	for { //这个for循环的作用是可以多次建立连接
		conn, err := listen.Accept() //请求建立连接，客户端未连接就会在这里一直等待
		if err != nil {
			fmt.Println("accept failed, err:", err)
			continue
		}
		fmt.Println("连接建立成功.....")
		go process(conn)
	}
}
func process(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}(conn)
	for {
		var buf [1024]byte
		n, err := conn.Read(buf[:]) //定义为切片 相当于buf[0:len(buf)]
		if err != nil {             //一直在读取,读取失败break
			log.Println("read from client failed, err:", err)
			break
		}
		log.Println("8002收到client端发来的数据")
		ParseDTUData(buf[:], n)
	}
}
