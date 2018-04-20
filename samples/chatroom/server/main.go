package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"tcpsock.v2/samples/chatroom/server/cfgmgr"
	"tcpsock.v2/samples/chatroom/server/gamemgr"
	"tcpsock.v2/samples/chatroom/server/listensock"
	. "tcpsock.v2/samples/chatroom/server/msgnode"
)

var (
	shutdown  = make(chan bool, 1)
	exitChan  = make(chan struct{})
	waitGroup = &sync.WaitGroup{}

	msgChan = make(chan *MsgNode, 5)
)

func init() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		shutdown <- true
	}()
}

func main() {
	fmt.Println("chatroom server version 1.0.0, copyright (c) 2017~2018 ecofast")

	setup()
	serve()
	cleanup()
}

func setup() {
	cfgmgr.Setup()
	gamemgr.Setup()
	listensock.Setup()
}

func cleanup() {
	//
}

func serve() {
	log.Println("=====chatroom service start=====")
	waitGroup.Add(2)
	go gamemgr.Run(exitChan, waitGroup, msgChan)
	go listensock.Run(exitChan, waitGroup, msgChan)
	<-shutdown
	close(exitChan)
	waitGroup.Wait()
	log.Println("=====chatroom service end=====")
}
