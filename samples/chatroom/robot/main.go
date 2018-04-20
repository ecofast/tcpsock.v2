package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"

	"tcpsock.v2/samples/chatroom/protocol"
)

var (
	shutdown = make(chan bool, 1)

	svrAddr  string
	robotNum int
	roomID   int
)

func init() {
	rand.Seed(time.Now().UnixNano())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		shutdown <- true
	}()
}

func parseFlag() {
	flag.StringVar(&svrAddr, "s", "127.0.0.1:12321", "server address")
	flag.IntVar(&robotNum, "n", 200, "num of robot")
	flag.IntVar(&roomID, "r", 0, "room id")
	flag.Parse()
	fmt.Printf("server addr: %s\n", svrAddr)
	fmt.Printf("num of robot: %d\n", robotNum)
	fmt.Printf("room id: %d\n", roomID)
}

func main() {
	parseFlag()

	for i := 0; i < robotNum; i++ {
		go func(idx int) {
			cli := newTcpClient(svrAddr, idx)
			cli.Open()
			cli.genUserName()
			cli.Write(protocol.NewPacket(protocol.PT_NORMAL, protocol.CM_ENTERROOM, uint16(roomID), nil).Bytes())
			cli.chat()
			<-shutdown
			go cli.Write(protocol.NewPacket(protocol.PT_NORMAL, protocol.CM_EXITROOM, 0, nil).Bytes())
			time.Sleep(time.Second)
			cli.Close()
		}(i)
	}

	<-shutdown
}

func str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}
