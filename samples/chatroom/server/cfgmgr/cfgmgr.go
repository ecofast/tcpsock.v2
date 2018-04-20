package cfgmgr

import (
	"log"
	"os"

	"github.com/ecofast/rtl/inifiles"
	"github.com/ecofast/rtl/sysutils"
)

type config struct {
	clientListenPort int
	snapshotLogIntv  int // secs
}

var (
	cfg *config
)

func Setup() {
	iniName := sysutils.ChangeFileExt(os.Args[0], ".ini")
	ini := inifiles.New(iniName, false)
	cfg = &config{
		clientListenPort: ini.ReadInt("setup", "ClientListenPort", 12321),
		snapshotLogIntv:  ini.ReadInt("setup", "SnapshotLogIntv", 0),
	}
	if cfg.clientListenPort <= 1024 || cfg.clientListenPort >= 65536 || cfg.snapshotLogIntv < 0 {
		panic("invalid configuration!")
	}
	log.Println("configuration has been loaded successfully")
}

func ClientListenPort() int {
	return cfg.clientListenPort
}

func SnapshotLogIntv() int {
	return cfg.snapshotLogIntv
}
