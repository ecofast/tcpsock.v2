package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unsafe"

	. "github.com/ecofast/rtl/sysutils"
	"tcpsock.v2/samples/chatroom/protocol"
)

const (
	charTableLen = 62
	charTable    = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var (
	shutdown = make(chan bool, 1)

	cli      *client
	userName [protocol.SizeOfUserName]byte
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

func main() {
	if len(os.Args) == 1 {
		panic("invalid server addr")
	}
	cli = newTcpClient(os.Args[1])
	cli.Open()
	genUserName()
	go input()
	<-shutdown
	if cli.roomID != 0xFF {
		go cli.Write(protocol.NewPacket(protocol.PT_NORMAL, protocol.CM_EXITROOM, 0, nil).Bytes())
	}
	time.Sleep(time.Second)
	cli.Close()
}

func genUserName() {
	for i := 0; i < protocol.SizeOfUserName; i++ {
		userName[i] = charTable[rand.Intn(charTableLen)]
	}
	fmt.Println("your random name is:", string(userName[:]))

	cli.Write(protocol.NewPacket(protocol.PT_NORMAL, protocol.CM_IDENTITY, 0, userName[:]).Bytes())
}

func input() {
	reader := bufio.NewReader(os.Stdin)
	for {
		if cli == nil {
			break
		}
		if s, err := reader.ReadString('\n'); err == nil {
			s = strings.TrimSpace(s)
			if strings.HasPrefix(s, "rooms?") {
				cli.Write(protocol.NewPacket(protocol.PT_NORMAL, protocol.CM_REQROOMLIST, 0, nil).Bytes())
			} else if strings.HasPrefix(s, "enter:") {
				if cli.roomID == 0xFF {
					idx := strings.Index(s, ":")
					cli.Write(protocol.NewPacket(protocol.PT_NORMAL, protocol.CM_ENTERROOM, uint16(StrToInt(s[idx+1:])), nil).Bytes())
				}
			} else {
				cli.Write(protocol.NewPacket(protocol.PT_NORMAL, protocol.CM_CHAT, 0, str2bytes(s)).Bytes())
			}
			continue
		}
		break
	}
}

func str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}
