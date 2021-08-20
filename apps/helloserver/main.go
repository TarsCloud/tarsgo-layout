package main

import (
	"github.com/TarsCloud/TarsGo/tars"
	"github.com/tarscloud/gopractice/apps/helloserver/logic"
	"github.com/tarscloud/gopractice/apps/helloserver/proto/stub/Base"
	"github.com/tarscloud/gopractice/common/initserver"
)

func main() {
	// Init server
	cfg := tars.GetServerConfig()
	if err := initserver.NewOption().
		// WithRemoteConf(cfg.Server+".json", config.Init).
		DoInit(); err != nil {
		panic(err)
	}

	// Add servant
	imp := new(logic.ServerImp)
	app := new(Base.Main)
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".MainObj")

	// Run application
	tars.Run()
}
