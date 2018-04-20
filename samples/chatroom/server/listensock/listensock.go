package listensock

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"tcpsock.v2"
	"tcpsock.v2/samples/chatroom/server/cfgmgr"
	"tcpsock.v2/samples/chatroom/server/clientsock"
	. "tcpsock.v2/samples/chatroom/server/msgnode"
)

type chatServer struct {
	*tcpsock.TcpServer
	cliChan chan<- *MsgNode
}

var (
	chatSvr *chatServer
	ticker  *time.Ticker
)

func Setup() {
	fmt.Printf("client listen port: %d\n", cfgmgr.ClientListenPort())
}

func Run(exitChan chan struct{}, waitGroup *sync.WaitGroup, cliChan chan<- *MsgNode) {
	defer waitGroup.Done()

	chatSvr = newChatServer(fmt.Sprintf(":%d", cfgmgr.ClientListenPort()), cliChan)
	chatSvr.Serve()

	intv := cfgmgr.SnapshotLogIntv()
	if intv > 0 {
		ticker = time.NewTicker(time.Duration(intv) * time.Second)
		go func() {
			for range ticker.C {
				log.Printf("Number of Concurrent Users: %d\n", chatSvr.Count())
			}
		}()
	}

	<-exitChan

	if ticker != nil {
		ticker.Stop()
	}
	chatSvr.Close()
}

func newChatServer(addr string, cliChan chan<- *MsgNode) *chatServer {
	svr := &chatServer{}
	svr.TcpServer = tcpsock.NewTcpServer(addr, svr.onConnect, svr.onDisconnect, svr.onCheckIP)
	svr.cliChan = cliChan
	return svr
}

func (self *chatServer) onConnect(conn *tcpsock.TcpConn) tcpsock.TcpSession {
	cli := clientsock.New(conn.ID(), conn.Write, conn.Close, self.cliChan)
	return cli
}

func (self *chatServer) onDisconnect(conn *tcpsock.TcpConn) {
	//
}

func (self *chatServer) onCheckIP(ip net.Addr) bool {
	return true
}
