package router

// golang实现带有心跳检测的tcp长连接
// server

import (
	"fmt"
	"net"
	"time"
)

var (
	Req_REGISTER byte = 1 // 1 --- c register cid
	Res_REGISTER byte = 2 // 2 --- s response

	Req_HEARTBEAT byte = 3 // 3 --- s send heartbeat req
	Res_HEARTBEAT byte = 4 // 4 --- c send heartbeat res

	Req byte = 5 // 5 --- cs send data
	Res byte = 6 // 6 --- cs send ack
)

var Dch chan bool
var Rch chan []byte
var Wch chan []byte

var Server string

func HeartBeatRoute(server string) {
	// config := struct {
	// 	Section struct {
	//     	Server    string
	// 	}
	// }{}
	// err := gcfg.ReadFileInto(&config, "./remote.conf")
	// if err != nil {
	//     fmt.Printf("Failed to parse config file: %s", err)
	// }
	// Server = config.Section.Server
	// Server = ""
	Server = server
	Dch = make(chan bool)
	Rch = make(chan []byte)
	Wch = make(chan []byte)
	connectTimes := 0

	for {
		if connected := connect(); connected {
			connectTimes = 0
			<-Dch
			fmt.Println("关闭连接")
		} else {
			connectTimes++
			time.Sleep(time.Second * 5)
		}
		if connectTimes > 20 {
			break
		}
	}
}

func TestMoudle() {
	fmt.Println("test mouble success")
}

func connect() bool {
	addr, _ := net.ResolveTCPAddr("tcp", Server)
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Println("连接服务端失败:", err.Error())
		return false
	}
	fmt.Println("已连接服务器")
	//defer conn.Close()
	go Handler(conn)
	return true
}

func Handler(conn *net.TCPConn) {
	// 直到register ok
	data := make([]byte, 128)
	for {
		conn.Write([]byte{Req_REGISTER, '#', '2'})
		conn.Read(data)
		//		fmt.Println(string(data))
		if data[0] == Res_REGISTER {
			break
		}
	}
	//	fmt.Println("i'm register")
	go RHandler(conn)
	go WHandler(conn)
	go Work()
}

func RHandler(conn *net.TCPConn) {

	for {
		err := conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		if err != nil {
			fmt.Println(err)
		}

		// 心跳包,回复ack
		data := make([]byte, 128)
		i, _ := conn.Read(data)
		if i == 0 {
			fmt.Println(time.Now().Format(time.ANSIC))
			conn.Close()
			Dch <- true
			return
		}
		if data[0] == Req_HEARTBEAT {
			//fmt.Println("recv ht pack")
			conn.Write([]byte{Res_HEARTBEAT, '#', 'h'})
			//fmt.Println("send ht pack ack")
		} else if data[0] == Req {
			fmt.Println("recv data pack")
			Rch <- data[2:]
			conn.Write([]byte{Res, '#'})
		}
	}
}

func WHandler(conn net.Conn) {
	for msg := range Wch {
		fmt.Println("send data after: " + string(msg[1:]))
		conn.Write(msg)
	}
}

func Work() {
	magicPkg := make([]byte, 102)
	mac := []byte{0xd0, 0x50, 0x99, 0x14, 0xe3, 0x12}
	for i := 0; i < 6; i++ {
		magicPkg[i] = 0xFF
	}
	for i := 0; i < 16; i++ {
		for j := 0; j < 6; j++ {
			magicPkg[6+i*6+j] = mac[j]
		}
	}
	for {
		msg := <-Rch
		fmt.Println("work recv " + string(msg))
		con, err := net.Dial("udp", "192.168.0.255:9")
		if err != nil {
			fmt.Println(err)
			continue
		}
		con.Write(magicPkg)
		con.Close()
	}
}
