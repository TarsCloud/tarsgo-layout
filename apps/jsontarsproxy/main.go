package main

import (
	"github.com/tarscloud/gopractice/apps/jsontarsproxy/config"
	"github.com/tarscloud/gopractice/apps/jsontarsproxy/logic"
	"github.com/tarscloud/gopractice/common/initserver"

	"github.com/TarsCloud/TarsGo/tars"
)

func main() {
	// init server
	cfg := tars.GetServerConfig()
	if err := initserver.NewOption().
		WithRemoteConf(cfg.Server+".json", config.Init).
		DoInit(); err != nil {
		panic(err)
	}

	// add servant
	mux := &tars.TarsHttpMux{}
	mux.HandleFunc("/", logic.HandlerFunc)
	tars.AddHttpServant(mux, cfg.App+"."+cfg.Server+".HttpObj")

	// run application
	tars.Run()
}
