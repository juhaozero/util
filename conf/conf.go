package conf

import (
	"flag"
)

var (
	ConfigPath string
)

func init() {
	flag.StringVar(&ConfigPath, "c", "", "default config path")
	ServerConfig = &ServerConf{}
	//flag.StringVar(&App, "app", "", "default app")
}

// var ServerConfig *GameServerConf
var ServerConfig *ServerConf

type ServerConf struct {
	App
	Log
}

type App struct {
	AppId    string
	AppName  string
	HttpAddr int32
	Test     bool
}

type Log struct {
	LogPath string
	IsDebug bool
	LogTime int32
}
